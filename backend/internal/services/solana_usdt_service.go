package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"certhub-backend/internal/config"
)

var (
	ErrUSDTDisabled         = errors.New("USDT 支付未启用")
	ErrUSDTConfigIncomplete = errors.New("USDT 支付配置不完整")
	ErrTxNotFound           = errors.New("链上交易不存在或尚未确认")
	ErrTxFailed             = errors.New("链上交易执行失败")
	ErrTxNotUSDTTransfer    = errors.New("交易中未找到向收款地址转入的 USDT")
	ErrTxAmountMismatch     = errors.New("链上到账金额与充值金额不匹配")
	ErrTxHashInvalid        = errors.New("交易哈希无效")
)

type solanaRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type solanaRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type solanaTokenAmount struct {
	Amount   string   `json:"amount"`
	Decimals int      `json:"decimals"`
	UIAmount *float64 `json:"uiAmount"`
}

type solanaTokenBalance struct {
	AccountIndex  int               `json:"accountIndex"`
	Mint          string            `json:"mint"`
	Owner         string            `json:"owner"`
	UITokenAmount solanaTokenAmount `json:"uiTokenAmount"`
}

type solanaTxMeta struct {
	Err               interface{}          `json:"err"`
	PreTokenBalances  []solanaTokenBalance `json:"preTokenBalances"`
	PostTokenBalances []solanaTokenBalance `json:"postTokenBalances"`
}

type solanaTxResult struct {
	Meta *solanaTxMeta `json:"meta"`
}

type solanaGetTxResponse struct {
	Result *solanaTxResult `json:"result"`
	Error  *solanaRPCError `json:"error"`
}

// VerifySolanaUSDTTransfer checks that txHash is a successful Solana USDT transfer
// to the configured recipient for at least the expected CNY amount (within tolerance).
func VerifySolanaUSDTTransfer(ctx context.Context, txHash string, expectedCNY float64) (receivedUSDT float64, err error) {
	cfg := config.C.Payment.USDT
	if !cfg.Enabled {
		return 0, ErrUSDTDisabled
	}
	if cfg.Recipient == "" || cfg.Mint == "" || cfg.RPCURL == "" {
		return 0, ErrUSDTConfigIncomplete
	}
	txHash = strings.TrimSpace(txHash)
	if txHash == "" || len(txHash) < 64 {
		return 0, ErrTxHashInvalid
	}

	tx, err := fetchSolanaTransaction(ctx, cfg.RPCURL, txHash)
	if err != nil {
		return 0, err
	}
	if tx == nil || tx.Meta == nil {
		return 0, ErrTxNotFound
	}
	if tx.Meta.Err != nil {
		return 0, ErrTxFailed
	}

	receivedUSDT, ok := calcUSDTReceivedByOwner(tx.Meta, cfg.Recipient, cfg.Mint)
	if !ok || receivedUSDT <= 0 {
		return 0, ErrTxNotUSDTTransfer
	}

	cnyPerUsdt, err := fetchCnyPerUsdt(ctx)
	if err != nil || cnyPerUsdt <= 0 {
		cnyPerUsdt = cfg.FallbackCnyPerUsdt
	}
	expectedUSDT := expectedCNY / cnyPerUsdt
	tolerance := cfg.AmountTolerancePercent / 100
	minAcceptable := expectedUSDT*(1-tolerance) - 0.01
	if minAcceptable < 0 {
		minAcceptable = 0
	}

	if receivedUSDT+1e-9 < minAcceptable {
		return receivedUSDT, fmt.Errorf("%w: 期望约 %.4f USDT，实际 %.4f USDT", ErrTxAmountMismatch, expectedUSDT, receivedUSDT)
	}
	return receivedUSDT, nil
}

func fetchSolanaTransaction(ctx context.Context, rpcURL, signature string) (*solanaTxResult, error) {
	reqBody := solanaRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getTransaction",
		Params: []interface{}{
			signature,
			map[string]interface{}{
				"encoding":                       "jsonParsed",
				"commitment":                     "confirmed",
				"maxSupportedTransactionVersion": 0,
			},
		},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求 Solana RPC 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Solana RPC 返回 HTTP %d", resp.StatusCode)
	}

	var rpcResp solanaGetTxResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("解析 Solana 交易失败: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("Solana RPC 错误: %s", rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

func calcUSDTReceivedByOwner(meta *solanaTxMeta, owner, mint string) (float64, bool) {
	pre := tokenAmountByOwnerMint(meta.PreTokenBalances, owner, mint)
	post := tokenAmountByOwnerMint(meta.PostTokenBalances, owner, mint)
	delta := post - pre
	if delta <= 0 {
		return 0, false
	}
	return delta, true
}

func tokenAmountByOwnerMint(balances []solanaTokenBalance, owner, mint string) float64 {
	var total float64
	for _, b := range balances {
		if b.Owner != owner || b.Mint != mint {
			continue
		}
		if b.UITokenAmount.UIAmount != nil {
			total += *b.UITokenAmount.UIAmount
			continue
		}
		if amt, err := parseTokenAmountString(b.UITokenAmount.Amount, b.UITokenAmount.Decimals); err == nil {
			total += amt
		}
	}
	return total
}

func parseTokenAmountString(amount string, decimals int) (float64, error) {
	if amount == "" {
		return 0, errors.New("empty amount")
	}
	var raw float64
	if _, err := fmt.Sscan(amount, &raw); err != nil {
		return 0, err
	}
	if decimals <= 0 {
		return raw, nil
	}
	return raw / math.Pow10(decimals), nil
}

func fetchCnyPerUsdt(ctx context.Context) (float64, error) {
	if rate, err := fetchUSDTCnyFromCoinGecko(ctx); err == nil && rate > 0 {
		return rate, nil
	}
	return fetchUSDCnyRate(ctx)
}

func fetchUSDTCnyFromCoinGecko(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=cny",
		nil,
	)
	if err != nil {
		return 0, err
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coingecko http %d", resp.StatusCode)
	}
	var data struct {
		Tether struct {
			CNY float64 `json:"cny"`
		} `json:"tether"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	if data.Tether.CNY <= 0 {
		return 0, errors.New("coingecko rate missing")
	}
	return data.Tether.CNY, nil
}

func fetchUSDCnyRate(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.exchangerate-api.com/v4/latest/USD",
		nil,
	)
	if err != nil {
		return 0, err
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("exchangerate http %d", resp.StatusCode)
	}
	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	rate := data.Rates["CNY"]
	if rate <= 0 {
		return 0, errors.New("usd/cny rate missing")
	}
	return rate, nil
}
