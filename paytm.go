// Package paytm provides a Go SDK for the Paytm Payment Gateway API.
//
// It supports initiating transactions, checking payment status, initiating refunds,
// and checking refund status. The SDK handles checksum generation and verification
// using the Paytm signature algorithm (SHA256 + AES-CBC), matching the official
// Paytm Node.js SDK.
//
// Usage:
//
//	client := paytm.NewClient("YOUR_MID", "YOUR_KEY", "WEBSTAGING", paytm.EnvStaging)
//	resp, err := client.InitiateTransaction(ctx, paytm.InitiateTransactionRequest{
//	    ChannelID: paytm.ChannelWeb,
//	    OrderID:   "ORDER_001",
//	    TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
//	    UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
//	})
package paytm

import (
	"context"
	"net/http"
	"time"

	"github.com/KriaaCompany/paytm-go-sdk/internal/checksum"
)

// Client is the Paytm Payment Gateway API client.
type Client struct {
	mid         string
	merchantKey string
	website     string
	env         Environment
	httpClient  *http.Client
	clientID    string // optional, sent in request head
	callbackURL string // optional default callback URL
}

// NewClient creates a new Paytm API client.
//
// Parameters:
//   - mid: Merchant ID provided by Paytm
//   - merchantKey: Merchant key for checksum signing (must be 16, 24, or 32 bytes)
//   - website: Website name (e.g., "WEBSTAGING" for staging, "DEFAULT" for production)
//   - env: Environment (EnvStaging or EnvProduction)
//   - opts: Optional functional options
func NewClient(mid, merchantKey, website string, env Environment, opts ...Option) *Client {
	c := &Client{
		mid:         mid,
		merchantKey: merchantKey,
		website:     website,
		env:         env,
		httpClient: &http.Client{
			Timeout: 80 * time.Second, // matches Node SDK default readTimeout
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// initiateTransactionBody is the JSON body for the initiateTransaction API.
// channelId lives in the head, not the body (matches Node SDK InitiateTransactionRequestBody).
type initiateTransactionBody struct {
	RequestType            string        `json:"requestType"`
	MID                    string        `json:"mid"`
	OrderID                string        `json:"orderId"`
	WebsiteName            string        `json:"websiteName,omitempty"`
	TxnAmount              Money         `json:"txnAmount"`
	UserInfo               UserInfo      `json:"userInfo"`
	PaytmSsoToken          string        `json:"paytmSsoToken,omitempty"`
	EnablePaymentMode      []PaymentMode `json:"enablePaymentMode,omitempty"`
	DisablePaymentMode     []PaymentMode `json:"disablePaymentMode,omitempty"`
	PromoCode              string        `json:"promoCode,omitempty"`
	CallbackURL            string        `json:"callbackUrl,omitempty"`
	Goods                  []GoodsInfo   `json:"goods,omitempty"`
	ShippingInfo           []ShippingInfo `json:"shippingInfo,omitempty"`
	ExtendInfo             *ExtendInfo   `json:"extendInfo,omitempty"`
	EMIOption              string        `json:"emiOption,omitempty"`
	CardTokenRequired      string        `json:"cardTokenRequired,omitempty"`
	CartValidationRequired string        `json:"cartValidationRequired,omitempty"`
}

// InitiateTransaction starts a new payment transaction and returns a transaction token.
func (c *Client) InitiateTransaction(ctx context.Context, req InitiateTransactionRequest) (*InitiateTransactionResponse, error) {
	if req.OrderID == "" {
		return nil, newError(ErrMissingMandatoryParams, "orderId is required")
	}
	if req.TxnAmount.Value == "" {
		return nil, newError(ErrMissingMandatoryParams, "txnAmount.value is required")
	}
	if req.UserInfo.CustID == "" {
		return nil, newError(ErrMissingMandatoryParams, "userInfo.custId is required")
	}

	// Use request callbackUrl, falling back to client-level default
	callbackURL := req.CallbackURL
	if callbackURL == "" {
		callbackURL = c.callbackURL
	}

	body := initiateTransactionBody{
		RequestType:            "Payment",
		MID:                    c.mid,
		OrderID:                req.OrderID,
		WebsiteName:            c.website,
		TxnAmount:              req.TxnAmount,
		UserInfo:               req.UserInfo,
		PaytmSsoToken:          req.PaytmSsoToken,
		EnablePaymentMode:      req.EnablePaymentMode,
		DisablePaymentMode:     req.DisablePaymentMode,
		PromoCode:              req.PromoCode,
		CallbackURL:            callbackURL,
		Goods:                  req.Goods,
		ShippingInfo:           req.ShippingInfo,
		ExtendInfo:             req.ExtendInfo,
		EMIOption:              req.EMIOption,
		CardTokenRequired:      req.CardTokenRequired,
		CartValidationRequired: req.CartValidationRequired,
	}

	url := c.env.initiateTransactionURL(c.mid, req.OrderID)

	var resp InitiateTransactionResponse
	if err := c.doRequest(ctx, url, req.ChannelID, req.WorkFlow, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// paymentStatusBody is the JSON body for the getPaymentStatus API.
type paymentStatusBody struct {
	MID     string `json:"mid"`
	OrderID string `json:"orderId"`
	TxnType string `json:"txnType,omitempty"`
}

// GetPaymentStatus retrieves the current status of a payment transaction.
func (c *Client) GetPaymentStatus(ctx context.Context, req PaymentStatusRequest) (*PaymentStatusResponse, error) {
	if req.OrderID == "" {
		return nil, newError(ErrMissingMandatoryParams, "orderId is required")
	}

	body := paymentStatusBody{
		MID:     c.mid,
		OrderID: req.OrderID,
		TxnType: req.TxnType,
	}

	var resp PaymentStatusResponse
	if err := c.doRequest(ctx, c.env.paymentStatusURL(), "", "", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// refundBody is the JSON body for the async refund API.
type refundBody struct {
	MID                  string        `json:"mid"`
	OrderID              string        `json:"orderId"`
	RefID                string        `json:"refId"`
	TxnID                string        `json:"txnId"`
	TxnType              string        `json:"txnType"`
	RefundAmount         string        `json:"refundAmount"`
	Comments             string        `json:"comments,omitempty"`
	PreferredDestination string        `json:"preferredDestination,omitempty"`
	RequestID            string        `json:"requestId,omitempty"`
	SubwalletAmount      []interface{} `json:"subwalletAmount,omitempty"`
	ExtraParamsMap       map[string]interface{} `json:"extraParamsMap,omitempty"`
}

// InitiateRefund initiates a refund for a completed transaction.
func (c *Client) InitiateRefund(ctx context.Context, req RefundRequest) (*RefundResponse, error) {
	if req.OrderID == "" {
		return nil, newError(ErrMissingMandatoryParams, "orderId is required")
	}
	if req.RefID == "" {
		return nil, newError(ErrMissingMandatoryParams, "refId is required")
	}
	if req.TxnID == "" {
		return nil, newError(ErrMissingMandatoryParams, "txnId is required")
	}
	if req.TxnType == "" {
		return nil, newError(ErrMissingMandatoryParams, "txnType is required")
	}
	if req.RefundAmount == "" {
		return nil, newError(ErrMissingMandatoryParams, "refundAmount is required")
	}

	// requestId defaults to refId (matches Node SDK getRequestId())
	requestID := req.RequestID
	if requestID == "" {
		requestID = req.RefID
	}

	body := refundBody{
		MID:                  c.mid,
		OrderID:              req.OrderID,
		RefID:                req.RefID,
		TxnID:                req.TxnID,
		TxnType:              req.TxnType,
		RefundAmount:         req.RefundAmount,
		Comments:             req.Comments,
		PreferredDestination: req.PreferredDestination,
		RequestID:            requestID,
		SubwalletAmount:      req.SubwalletAmount,
		ExtraParamsMap:       req.ExtraParamsMap,
	}

	var resp RefundResponse
	if err := c.doRequest(ctx, c.env.refundURL(), "", "", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// refundStatusBody is the JSON body for the refundStatus API.
type refundStatusBody struct {
	MID     string `json:"mid"`
	OrderID string `json:"orderId"`
	RefID   string `json:"refId"`
}

// GetRefundStatus retrieves the current status of a refund.
// Note: In production this API uses a different host (pgp-ite.paytm.in).
func (c *Client) GetRefundStatus(ctx context.Context, req RefundStatusRequest) (*RefundStatusResponse, error) {
	if req.OrderID == "" {
		return nil, newError(ErrMissingMandatoryParams, "orderId is required")
	}
	if req.RefID == "" {
		return nil, newError(ErrMissingMandatoryParams, "refId is required")
	}

	body := refundStatusBody{
		MID:     c.mid,
		OrderID: req.OrderID,
		RefID:   req.RefID,
	}

	var resp RefundStatusResponse
	if err := c.doRequest(ctx, c.env.refundStatusURL(), "", "", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VerifyChecksum verifies a Paytm checksum signature against a body string and merchant key.
// This is useful for verifying webhook payloads without a Client instance.
func VerifyChecksum(body, signature, merchantKey string) (bool, error) {
	return checksum.Verify(body, signature, merchantKey)
}
