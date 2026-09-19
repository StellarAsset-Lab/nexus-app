package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RPCHealth is the getHealth result shape per the official Stellar RPC API
// reference (developers.stellar.org/docs/data/apis/rpc/api-reference/methods/getHealth).
type RPCHealth struct {
	Status                string `json:"status"`
	LatestLedger          uint32 `json:"latestLedger"`
	OldestLedger          uint32 `json:"oldestLedger"`
	LedgerRetentionWindow uint32 `json:"ledgerRetentionWindow"`
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	Result *RPCHealth `json:"result"`
	Error  *rpcError  `json:"error"`
}

// CheckRPCHealth calls the Stellar RPC getHealth method directly over
// net/http. This is a simple stdlib JSON-RPC call, not a Soroban contract
// interaction, so it deliberately does not depend on the Stellar Go SDK.
func CheckRPCHealth(ctx context.Context, rpcURL string) (RPCHealth, error) {
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: 1, Method: "getHealth"})
	if err != nil {
		return RPCHealth{}, fmt.Errorf("marshal getHealth request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(body))
	if err != nil {
		return RPCHealth{}, fmt.Errorf("build getHealth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return RPCHealth{}, fmt.Errorf("getHealth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RPCHealth{}, fmt.Errorf("getHealth returned HTTP %d", resp.StatusCode)
	}

	var parsed rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return RPCHealth{}, fmt.Errorf("decode getHealth response: %w", err)
	}

	if parsed.Error != nil {
		return RPCHealth{}, fmt.Errorf("getHealth rpc error %d: %s", parsed.Error.Code, parsed.Error.Message)
	}
	if parsed.Result == nil {
		return RPCHealth{}, fmt.Errorf("getHealth response missing result")
	}

	return *parsed.Result, nil
}
