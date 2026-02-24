package paytm

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KriaaCompany/paytm-go-sdk/internal/checksum"
)

const (
	pathInitiateTransaction = "/theia/api/v1/initiateTransaction"
	pathPaymentStatus       = "/merchant-status/api/v1/getPaymentStatus"
	pathRefund              = "/refund/api/v1/async/refund"
	pathRefundStatus        = "/refund/api/v1/refundStatus"
)

// requestHead represents the "head" section of a Paytm API request.
type requestHead struct {
	Version          string `json:"version,omitempty"`
	ChannelID        string `json:"channelId,omitempty"`
	RequestTimestamp string `json:"requestTimestamp,omitempty"`
	Signature        string `json:"signature"`
}

// apiRequest wraps the head and body of a Paytm API request.
type apiRequest struct {
	Head requestHead     `json:"head"`
	Body json.RawMessage `json:"body"`
}

// apiResponse wraps the head and body of a Paytm API response.
type apiResponse struct {
	Head struct {
		Signature string `json:"signature"`
	} `json:"head"`
	Body json.RawMessage `json:"body"`
}

// doRequest builds, signs, sends an API request, verifies the response signature, and unmarshals the body.
func (c *Client) doRequest(ctx context.Context, path string, queryMID string, queryOrderID string, bodyObj interface{}, result interface{}) error {
	// Marshal body
	bodyBytes, err := json.Marshal(bodyObj)
	if err != nil {
		return newError(ErrJSONConversionFailed, fmt.Sprintf("failed to marshal request body: %v", err))
	}

	// Generate signature over the body JSON
	signature, err := checksum.Generate(string(bodyBytes), c.merchantKey)
	if err != nil {
		return newError(ErrSignatureValidationFailed, fmt.Sprintf("failed to generate signature: %v", err))
	}

	// Build the envelope
	envelope := apiRequest{
		Head: requestHead{
			Version:          "v1",
			RequestTimestamp: time.Now().UTC().Format(time.RFC3339),
			Signature:        signature,
		},
		Body: bodyBytes,
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return newError(ErrJSONConversionFailed, fmt.Sprintf("failed to marshal request envelope: %v", err))
	}

	// Build URL with query params
	url := c.env.baseURL() + path + "?mid=" + queryMID + "&orderId=" + queryOrderID

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(envelopeBytes))
	if err != nil {
		return newError(ErrAPICallFailed, fmt.Sprintf("failed to create HTTP request: %v", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", generateRequestID())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return newError(ErrAPICallFailed, fmt.Sprintf("HTTP request failed: %v", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return newError(ErrAPICallFailed, fmt.Sprintf("failed to read response body: %v", err))
	}

	if resp.StatusCode != http.StatusOK {
		return newErrorWithRaw(ErrAPICallFailed, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), string(respBody))
	}

	// Parse response envelope
	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return newErrorWithRaw(ErrJSONConversionFailed, fmt.Sprintf("failed to parse response: %v", err), string(respBody))
	}

	// Verify response signature
	if apiResp.Head.Signature != "" {
		valid, err := checksum.Verify(string(apiResp.Body), apiResp.Head.Signature, c.merchantKey)
		if err != nil {
			return newError(ErrSignatureValidationFailed, fmt.Sprintf("failed to verify response signature: %v", err))
		}
		if !valid {
			return newErrorWithRaw(ErrSignatureValidationFailed, "response signature verification failed", string(respBody))
		}
	}

	// Unmarshal the body into the result
	if err := json.Unmarshal(apiResp.Body, result); err != nil {
		return newErrorWithRaw(ErrJSONConversionFailed, fmt.Sprintf("failed to unmarshal response body: %v", err), string(respBody))
	}

	return nil
}

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "GO-SDK:" + sdkVersion + ":" + hex.EncodeToString(b)
}
