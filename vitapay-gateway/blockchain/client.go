package blockchain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps the Cosmos REST API for VitaCoin chain queries.
type Client struct {
	restEndpoint string
	httpClient   *http.Client
}

// NewClient returns a blockchain Client pointed at the given REST endpoint.
func NewClient(restEndpoint string) *Client {
	return &Client{
		restEndpoint: restEndpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TxResponse represents a simplified Cosmos SDK tx response.
type TxResponse struct {
	TxHash  string `json:"txhash"`
	Code    int    `json:"code"`
	RawLog  string `json:"raw_log"`
	Height  string `json:"height"`
	GasUsed string `json:"gas_used"`
}

// MerchantInfo represents a merchant registered on-chain.
type MerchantInfo struct {
	Address      string `json:"address"`
	BusinessName string `json:"business_name"`
	IsActive     bool   `json:"is_active"`
}

// BalanceResponse represents a coin balance.
type BalanceResponse struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// txApiResponse is the raw Cosmos REST response envelope for a single tx.
type txApiResponse struct {
	TxResponse TxResponse `json:"tx_response"`
}

// balanceApiResponse is the raw Cosmos REST response for balance query.
type balanceApiResponse struct {
	Balance BalanceResponse `json:"balance"`
}

// merchantApiResponse is the VitaCoin module merchant query response.
type merchantApiResponse struct {
	Merchant MerchantInfo `json:"merchant"`
}

// GetTx fetches a transaction by hash from the chain REST API.
func (c *Client) GetTx(txHash string) (*TxResponse, error) {
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", c.restEndpoint, txHash)
	resp, err := c.httpClient.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("GetTx: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("GetTx: tx %s not found on chain", txHash)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetTx: unexpected status %d for tx %s", resp.StatusCode, txHash)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetTx: failed to read response: %w", err)
	}

	var envelope txApiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("GetTx: failed to parse response: %w", err)
	}

	return &envelope.TxResponse, nil
}

// GetMerchant queries the VitaCoin module for a merchant by bech32 address.
func (c *Client) GetMerchant(address string) (*MerchantInfo, error) {
	url := fmt.Sprintf("%s/vitacoin/vitacoin/merchant/%s", c.restEndpoint, address)
	resp, err := c.httpClient.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("GetMerchant: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("GetMerchant: merchant %s not found on chain", address)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetMerchant: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetMerchant: failed to read response: %w", err)
	}

	var result merchantApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("GetMerchant: failed to parse response: %w", err)
	}

	return &result.Merchant, nil
}

// GetBalance fetches the VITA balance for a bech32 address.
func (c *Client) GetBalance(address string) (string, error) {
	url := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=uvita", c.restEndpoint, address)
	resp, err := c.httpClient.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("GetBalance: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GetBalance: unexpected status %d for %s", resp.StatusCode, address)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("GetBalance: failed to read response: %w", err)
	}

	var result balanceApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("GetBalance: failed to parse response: %w", err)
	}

	return result.Balance.Amount, nil
}
