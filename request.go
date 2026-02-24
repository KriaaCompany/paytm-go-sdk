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
	"strconv"
	"time"

	"github.com/KriaaCompany/paytm-go-sdk/internal/checksum"
)

// requestHead represents the "head" section of a Paytm API request.
// Matches SecureRequestHeader in the Node SDK.
type requestHead struct {
	Version          string `json:"version,omitempty"`
	ChannelID        string `json:"channelId,omitempty"`
	RequestTimestamp string `json:"requestTimestamp,omitempty"`
	WorkFlow         string `json:"workFlow,omitempty"`
	ClientID         string `json:"clientId,omitempty"`
	Signature        string `json:"signature"`
}

// apiRequest is the outer JSON envelope sent to Paytm.
type apiRequest struct {
	Head requestHead     `json:"head"`
	Body json.RawMessage `json:"body"`
}

// apiResponse is the outer JSON envelope received from Paytm.
type apiResponse struct {
	Head struct {
		Signature string `json:"signature"`
	} `json:"head"`
	Body json.RawMessage `json:"body"`
}

// successStatuses are the result statuses for which Paytm includes a response signature.
// Matches Request.ts validateResponseSignature in Node SDK.
var successStatuses = map[string]bool{
	"S":           true,
	"PENDING":     true,
	"TXN_SUCCESS": true,
}

// doRequest signs and sends an API request, verifies the response signature (on success),
// and unmarshals the response body into result.
//
// fullURL is the complete endpoint URL including any query parameters.
// channelID is placed in the request head (empty string = omit).
func (c *Client) doRequest(ctx context.Context, fullURL string, channelID ChannelID, bodyObj interface{}, result interface{}) error {
	// Marshal body, skipping nil/zero values via omitempty tags
	bodyBytes, err := json.Marshal(bodyObj)
	if err != nil {
		return newError(ErrJSONConversionFailed, fmt.Sprintf("failed to marshal request body: %v", err))
	}

	// Sign the body JSON
	signature, err := checksum.Generate(string(bodyBytes), c.merchantKey)
	if err != nil {
		return newError(ErrSignatureValidationFailed, fmt.Sprintf("failed to generate signature: %v", err))
	}

	// Build request envelope with head + body
	envelope := apiRequest{
		Head: requestHead{
			Version:          headVersion,
			ChannelID:        string(channelID),
			RequestTimestamp: strconv.FormatInt(time.Now().UnixMilli(), 10),
			ClientID:         c.clientID,
			Signature:        signature,
		},
		Body: bodyBytes,
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return newError(ErrJSONConversionFailed, fmt.Sprintf("failed to marshal request envelope: %v", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(envelopeBytes))
	if err != nil {
		return newError(ErrAPICallFailed, fmt.Sprintf("failed to create HTTP request: %v", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", generateRequestID())

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

	// Verify response signature only for success statuses (matches Node SDK behaviour).
	if apiResp.Head.Signature != "" {
		var statusCheck struct {
			ResultInfo struct {
				ResultStatus string `json:"resultStatus"`
			} `json:"resultInfo"`
		}
		json.Unmarshal(apiResp.Body, &statusCheck) // best-effort

		if successStatuses[statusCheck.ResultInfo.ResultStatus] {
			valid, err := checksum.Verify(string(apiResp.Body), apiResp.Head.Signature, c.merchantKey)
			if err != nil {
				return newError(ErrSignatureValidationFailed, fmt.Sprintf("failed to verify response signature: %v", err))
			}
			if !valid {
				return newErrorWithRaw(ErrSignatureValidationFailed, "response signature verification failed", string(respBody))
			}
		}
	}

	// Unmarshal body into caller's result type
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
