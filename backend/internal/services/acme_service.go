package services

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"certhub-backend/internal/config"
)

// ACMEClient handles ACME protocol interactions
type ACMEClient struct {
	directoryURL string
	email        string
	privateKey   *ecdsa.PrivateKey
	directory    *ACMEDirectory
	accountURL   string
	httpClient   *http.Client
	mu           sync.Mutex
}

// ACMEDirectory represents ACME directory endpoints
type ACMEDirectory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
}

// ACMEOrder represents an ACME order
type ACMEOrder struct {
	Status         string   `json:"status"`
	Expires        string   `json:"expires"`
	Identifiers    []ACMEId `json:"identifiers"`
	Authorizations []string `json:"authorizations"`
	Finalize       string   `json:"finalize"`
	Certificate    string   `json:"certificate,omitempty"`
}

// ACMEId represents an ACME identifier
type ACMEId struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ACMEAuthorization represents an ACME authorization
type ACMEAuthorization struct {
	Status     string          `json:"status"`
	Expires    string          `json:"expires"`
	Identifier ACMEId          `json:"identifier"`
	Challenges []ACMEChallenge `json:"challenges"`
}

// ACMEChallenge represents an ACME challenge
type ACMEChallenge struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Status string `json:"status"`
	Token  string `json:"token"`
}

// PendingChallenge stores challenge info for later verification
type PendingChallenge struct {
	Domain         string
	Token          string
	KeyAuth        string
	DNSRecordName  string
	DNSRecordValue string
	ChallengeURL   string
	AuthzURL       string
	OrderURL       string
	CreatedAt      time.Time
}

var (
	acmeClient        *ACMEClient
	acmeClientOnce    sync.Once
	pendingChallenges = make(map[string]*PendingChallenge) // keyed by domain
	challengesMu      sync.RWMutex
)

// GetACMEClient returns singleton ACME client
func GetACMEClient() (*ACMEClient, error) {
	var initErr error
	acmeClientOnce.Do(func() {
		client := &ACMEClient{
			directoryURL: config.C.ACME.DirectoryURL,
			email:        config.C.ACME.Email,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}

		// Generate or load account key
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			initErr = fmt.Errorf("failed to generate account key: %w", err)
			return
		}
		client.privateKey = key

		// Fetch directory
		if err := client.fetchDirectory(); err != nil {
			initErr = fmt.Errorf("failed to fetch ACME directory: %w", err)
			return
		}

		// Register account
		if err := client.registerAccount(); err != nil {
			initErr = fmt.Errorf("failed to register ACME account: %w", err)
			return
		}

		acmeClient = client
	})

	if initErr != nil {
		return nil, initErr
	}
	return acmeClient, nil
}

// fetchDirectory fetches ACME directory
func (c *ACMEClient) fetchDirectory() error {
	resp, err := c.httpClient.Get(c.directoryURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dir ACMEDirectory
	if err := json.Unmarshal(body, &dir); err != nil {
		return err
	}
	c.directory = &dir
	return nil
}

// getNonce fetches a fresh nonce
func (c *ACMEClient) getNonce() (string, error) {
	resp, err := c.httpClient.Head(c.directory.NewNonce)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Header.Get("Replay-Nonce"), nil
}

// base64URLEncode encodes bytes to base64url
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// jwkThumbprint computes JWK thumbprint
func (c *ACMEClient) jwkThumbprint() string {
	pub := c.privateKey.PublicKey
	jwk := map[string]string{
		"crv": "P-256",
		"kty": "EC",
		"x":   base64URLEncode(pub.X.Bytes()),
		"y":   base64URLEncode(pub.Y.Bytes()),
	}
	// Canonical JSON (sorted keys)
	canonical := fmt.Sprintf(`{"crv":"%s","kty":"%s","x":"%s","y":"%s"}`,
		jwk["crv"], jwk["kty"], jwk["x"], jwk["y"])
	hash := sha256.Sum256([]byte(canonical))
	return base64URLEncode(hash[:])
}

// signJWS creates a JWS signed request
func (c *ACMEClient) signJWS(url string, payload interface{}, useKid bool) ([]byte, error) {
	nonce, err := c.getNonce()
	if err != nil {
		return nil, err
	}

	var payloadBytes []byte
	if payload == nil {
		payloadBytes = []byte{}
	} else {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	// Build protected header
	protected := map[string]interface{}{
		"alg":   "ES256",
		"nonce": nonce,
		"url":   url,
	}

	if useKid && c.accountURL != "" {
		protected["kid"] = c.accountURL
	} else {
		pub := c.privateKey.PublicKey
		protected["jwk"] = map[string]string{
			"crv": "P-256",
			"kty": "EC",
			"x":   base64URLEncode(pub.X.Bytes()),
			"y":   base64URLEncode(pub.Y.Bytes()),
		}
	}

	protectedBytes, err := json.Marshal(protected)
	if err != nil {
		return nil, err
	}

	protectedB64 := base64URLEncode(protectedBytes)
	payloadB64 := base64URLEncode(payloadBytes)

	// Sign
	signingInput := protectedB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, c.privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// ES256 signature: r || s (each 32 bytes)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	sig := make([]byte, 64)
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	sigB64 := base64URLEncode(sig)

	jws := map[string]string{
		"protected": protectedB64,
		"payload":   payloadB64,
		"signature": sigB64,
	}

	return json.Marshal(jws)
}

// registerAccount registers or fetches existing account
func (c *ACMEClient) registerAccount() error {
	payload := map[string]interface{}{
		"termsOfServiceAgreed": true,
		"contact":              []string{"mailto:" + c.email},
	}

	body, err := c.signJWS(c.directory.NewAccount, payload, false)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.directory.NewAccount, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("account registration failed: %s", string(respBody))
	}

	c.accountURL = resp.Header.Get("Location")
	return nil
}

// CreateOrder creates a new ACME order for DNS-01 challenge
func (c *ACMEClient) CreateOrder(domain string) (*PendingChallenge, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create order
	orderPayload := map[string]interface{}{
		"identifiers": []map[string]string{
			// RFC 8555: wildcard certificates must request "*.example.com" directly.
			{"type": "dns", "value": domain},
		},
	}

	body, err := c.signJWS(c.directory.NewOrder, orderPayload, true)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.directory.NewOrder, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("order creation failed: %s", string(respBody))
	}

	var order ACMEOrder
	if err := json.Unmarshal(respBody, &order); err != nil {
		return nil, err
	}

	orderURL := resp.Header.Get("Location")

	// Get authorization
	if len(order.Authorizations) == 0 {
		return nil, errors.New("no authorizations in order")
	}

	authzURL := order.Authorizations[0]
	authz, err := c.getAuthorization(authzURL)
	if err != nil {
		return nil, err
	}

	// Find DNS-01 challenge
	var dns01Challenge *ACMEChallenge
	for _, ch := range authz.Challenges {
		if ch.Type == "dns-01" {
			dns01Challenge = &ch
			break
		}
	}

	if dns01Challenge == nil {
		return nil, errors.New("DNS-01 challenge not available")
	}

	// Compute key authorization
	keyAuth := dns01Challenge.Token + "." + c.jwkThumbprint()

	// DNS TXT record value is base64url(sha256(keyAuth))
	hash := sha256.Sum256([]byte(keyAuth))
	dnsValue := base64URLEncode(hash[:])

	// DNS record name
	baseDomain := strings.TrimPrefix(domain, "*.")
	dnsName := "_acme-challenge." + baseDomain

	pending := &PendingChallenge{
		Domain:         domain,
		Token:          dns01Challenge.Token,
		KeyAuth:        keyAuth,
		DNSRecordName:  dnsName,
		DNSRecordValue: dnsValue,
		ChallengeURL:   dns01Challenge.URL,
		AuthzURL:       authzURL,
		OrderURL:       orderURL,
		CreatedAt:      time.Now(),
	}

	// Store for later verification
	challengesMu.Lock()
	pendingChallenges[domain] = pending
	challengesMu.Unlock()

	return pending, nil
}

// getAuthorization fetches authorization details
func (c *ACMEClient) getAuthorization(url string) (*ACMEAuthorization, error) {
	body, err := c.signJWS(url, nil, true)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var authz ACMEAuthorization
	if err := json.Unmarshal(respBody, &authz); err != nil {
		return nil, err
	}

	return &authz, nil
}

// GetPendingChallenge retrieves a pending challenge by domain
func GetPendingChallenge(domain string) (*PendingChallenge, bool) {
	challengesMu.RLock()
	defer challengesMu.RUnlock()
	ch, ok := pendingChallenges[domain]
	return ch, ok
}

// VerifyDNSChallenge notifies ACME server that DNS record is ready for verification
func (c *ACMEClient) VerifyDNSChallenge(pending *PendingChallenge) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Notify ACME server that challenge is ready
	payload := map[string]interface{}{}

	body, err := c.signJWS(pending.ChallengeURL, payload, true)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", pending.ChallengeURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("challenge verification failed: %s", string(respBody))
	}

	// Poll authorization until it's valid
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)
		authz, err := c.getAuthorization(pending.AuthzURL)
		if err != nil {
			return err
		}
		if authz.Status == "valid" {
			return nil
		}
		if authz.Status == "invalid" {
			return errors.New("DNS challenge validation failed")
		}
	}

	return errors.New("DNS challenge validation timeout")
}

// CertificateResult contains the issued certificate
type CertificateResult struct {
	Certificate string
	PrivateKey  string
	CA          string
	ExpiresAt   time.Time
}

// FinalizeOrder completes the order and retrieves the certificate
func (c *ACMEClient) FinalizeOrder(pending *PendingChallenge) (*CertificateResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate certificate key
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate key: %w", err)
	}

	// Create CSR (Certificate Signing Request)
	csrDER, err := createCSR(pending.Domain, certKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	// Get order to find finalize URL
	body, err := c.signJWS(pending.OrderURL, nil, true)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", pending.OrderURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var order ACMEOrder
	if err := json.Unmarshal(respBody, &order); err != nil {
		return nil, err
	}

	// Finalize order
	finalizePayload := map[string]interface{}{
		"csr": base64URLEncode(csrDER),
	}

	body, err = c.signJWS(order.Finalize, finalizePayload, true)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("POST", order.Finalize, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("order finalization failed: %s", string(respBody))
	}

	// Poll order until certificate is ready
	maxAttempts := 10
	var certURL string
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)
		body, err := c.signJWS(pending.OrderURL, nil, true)
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", pending.OrderURL, strings.NewReader(string(body)))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/jose+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		var order ACMEOrder
		if err := json.Unmarshal(respBody, &order); err != nil {
			continue
		}

		if order.Status == "valid" && order.Certificate != "" {
			certURL = order.Certificate
			break
		}
		if order.Status == "invalid" {
			return nil, errors.New("order failed")
		}
	}

	if certURL == "" {
		return nil, errors.New("certificate URL not found")
	}

	// Download certificate
	body, err = c.signJWS(certURL, nil, true)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("POST", certURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	certPEM, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert private key to PEM
	privateKeyPEM, err := privateKeyToPEM(certKey)
	if err != nil {
		return nil, err
	}

	// Parse certificate to get expiration
	expiresAt, err := parseCertExpiration(certPEM)
	if err != nil {
		// Default to 90 days if parsing fails
		expiresAt = time.Now().Add(90 * 24 * time.Hour)
	}

	return &CertificateResult{
		Certificate: string(certPEM),
		PrivateKey:  string(privateKeyPEM),
		CA:          "Let's Encrypt",
		ExpiresAt:   expiresAt,
	}, nil
}

// createCSR creates a Certificate Signing Request
func createCSR(domain string, key *ecdsa.PrivateKey) ([]byte, error) {
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: domain,
		},
		// SAN must match ACME order identifiers exactly.
		DNSNames: []string{domain},
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		return nil, err
	}

	return csrDER, nil
}

// privateKeyToPEM converts ECDSA private key to PEM format
func privateKeyToPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDER,
	}

	return pem.EncodeToMemory(block), nil
}

// parseCertExpiration parses certificate expiration date
func parseCertExpiration(certPEM []byte) (time.Time, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, errors.New("failed to decode PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}

	return cert.NotAfter, nil
}

// Ensure ACMEClient implements crypto.Signer if needed
var _ crypto.Signer = (*ecdsa.PrivateKey)(nil)
