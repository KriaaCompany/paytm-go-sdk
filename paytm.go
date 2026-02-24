// Package paytm provides a Go SDK for the Paytm Payment Gateway API.
//
// It supports initiating transactions, checking payment status, initiating refunds,
// and checking refund status. The SDK handles checksum generation and verification
// using the Paytm signature algorithm (SHA256 + AES-CBC).
//
// Usage:
//
//	client := paytm.NewClient("YOUR_MID", "YOUR_KEY", "WEBSTAGING", paytm.EnvStaging)
//	resp, err := client.InitiateTransaction(ctx, paytm.InitiateTransactionRequest{
//	    OrderID:   "ORDER_001",
//	    TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
//	    UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
//	    ChannelID: "WEB",
//	})
package paytm

import (
	"context"
	"net/http"
	"time"
)

// Client is the Paytm Payment Gateway API client.
type Client struct {
	mid         string
	merchantKey string
	website     string
	env         Environment
	httpClient  *http.Client
}

// NewClient creates a new Paytm API client.
//
// Parameters:
//   - mid: Merchant ID provided by Paytm
//   - merchantKey: Merchant key for checksum generation
//   - website: Website name (e.g., "WEBSTAGING" for staging, "DEFAULT" for production)
//   - env: Environment (EnvStaging or EnvProduction)
//   - opts: Optional functional options (WithHTTPClient, WithReadTimeout, WithConnectTimeout)
func NewClient(mid, merchantKey, website string, env Environment, opts ...Option) *Client {
	c := &Client{
		mid:         mid,
		merchantKey: merchantKey,
		website:     website,
		env:         env,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// initiateTransactionBody is the JSON body sent to the initiateTransaction API.
type initiateTransactionBody struct {
	MID       string `json:"mid"`
	OrderID   string `json:"orderId"`
	WebsiteName string `json:"websiteName,omitempty"`
	TxnAmount Money  `json:"txnAmount"`
	UserInfo  UserInfo `json:"userInfo"`
	ChannelID string `json:"channelId,omitempty"`
	CallbackURL string `json:"callbackUrl,omitempty"`
	PaymentModeFilter *PaymentModeFilter `json:"enablePaymentMode,omitempty"`
	PromoCode string `json:"promoCode,omitempty"`
}

// InitiateTransaction starts a new payment transaction.
// It returns a transaction token that can be used to open the Paytm payment page.
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

	body := initiateTransactionBody{
		MID:               c.mid,
		OrderID:           req.OrderID,
		WebsiteName:       c.website,
		TxnAmount:         req.TxnAmount,
		UserInfo:          req.UserInfo,
		ChannelID:         req.ChannelID,
		CallbackURL:       req.CallbackURL,
		PaymentModeFilter: req.PaymentModeFilter,
		PromoCode:         req.PromoCode,
	}

	var resp InitiateTransactionResponse
	if err := c.doRequest(ctx, pathInitiateTransaction, c.mid, req.OrderID, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// paymentStatusBody is the JSON body sent to the getPaymentStatus API.
type paymentStatusBody struct {
	MID     string `json:"mid"`
	OrderID string `json:"orderId"`
}

// GetPaymentStatus retrieves the status of a payment transaction.
func (c *Client) GetPaymentStatus(ctx context.Context, req PaymentStatusRequest) (*PaymentStatusResponse, error) {
	if req.OrderID == "" {
		return nil, newError(ErrMissingMandatoryParams, "orderId is required")
	}

	body := paymentStatusBody{
		MID:     c.mid,
		OrderID: req.OrderID,
	}

	var resp PaymentStatusResponse
	if err := c.doRequest(ctx, pathPaymentStatus, c.mid, req.OrderID, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// refundBody is the JSON body sent to the refund API.
type refundBody struct {
	MID          string `json:"mid"`
	OrderID      string `json:"orderId"`
	RefID        string `json:"refId"`
	TxnID        string `json:"txnId"`
	TxnType      string `json:"txnType"`
	RefundAmount string `json:"refundAmount"`
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
	if req.RefundAmount == "" {
		return nil, newError(ErrMissingMandatoryParams, "refundAmount is required")
	}

	body := refundBody{
		MID:          c.mid,
		OrderID:      req.OrderID,
		RefID:        req.RefID,
		TxnID:        req.TxnID,
		TxnType:      req.TxnType,
		RefundAmount: req.RefundAmount,
	}

	var resp RefundResponse
	if err := c.doRequest(ctx, pathRefund, c.mid, req.OrderID, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// refundStatusBody is the JSON body sent to the refundStatus API.
type refundStatusBody struct {
	MID     string `json:"mid"`
	OrderID string `json:"orderId"`
	RefID   string `json:"refId,omitempty"`
}

// GetRefundStatus retrieves the status of a refund.
func (c *Client) GetRefundStatus(ctx context.Context, req RefundStatusRequest) (*RefundStatusResponse, error) {
	if req.OrderID == "" {
		return nil, newError(ErrMissingMandatoryParams, "orderId is required")
	}

	body := refundStatusBody{
		MID:     c.mid,
		OrderID: req.OrderID,
		RefID:   req.RefID,
	}

	var resp RefundStatusResponse
	if err := c.doRequest(ctx, pathRefundStatus, c.mid, req.OrderID, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
