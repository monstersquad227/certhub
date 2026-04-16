package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	"certhub-backend/internal/config"
)

// EncryptAES encrypts plaintext using AES-GCM and returns base64 string.
func EncryptAES(plain string) (string, error) {
	key := []byte(config.C.AES.Key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts base64 string to plaintext using AES-GCM.
func DecryptAES(cipherB64 string) (string, error) {
	key := []byte(config.C.AES.Key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}


