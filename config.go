package paytm

import (
	"net/http"
	"time"
)

// Environment represents the Paytm API environment.
type Environment string

const (
	// EnvStaging points to the Paytm staging/sandbox environment.
	EnvStaging Environment = "STAGE"
	// EnvProduction points to the Paytm production environment.
	EnvProduction Environment = "PROD"
)

const (
	// Base URLs (match Paytm Node SDK MerchantProperties.ts)
	stagingBaseURL    = "https://securestage.paytmpayments.com"
	productionBaseURL = "https://secure.paytmpayments.com"

	// RefundStatus uses a different host in production.
	refundStatusProductionBaseURL = "https://pgp-ite.paytm.in"

	// API paths
	pathInitiateTransaction = "/theia/api/v1/initiateTransaction"
	pathPaymentStatus       = "/merchant-status/api/v1/getPaymentStatus"
	pathRefund              = "/refund/api/v1/async/refund"
	pathRefundStatus        = "/refund/api/v1/refundStatus"

	// headVersion matches LibraryConstants.VERSION in the Node SDK.
	headVersion = "v2"

	sdkVersion = "1.0.0"
)

// initiateTransactionURL returns the full URL including query parameters.
func (e Environment) initiateTransactionURL(mid, orderID string) string {
	return e.baseURL() + pathInitiateTransaction + "?mid=" + mid + "&orderId=" + orderID
}

// paymentStatusURL returns the payment status endpoint URL (no query params).
func (e Environment) paymentStatusURL() string {
	return e.baseURL() + pathPaymentStatus
}

// refundURL returns the refund endpoint URL (no query params).
func (e Environment) refundURL() string {
	return e.baseURL() + pathRefund
}

// refundStatusURL returns the refund status URL. Production uses a separate host.
func (e Environment) refundStatusURL() string {
	if e == EnvProduction {
		return refundStatusProductionBaseURL + pathRefundStatus
	}
	return stagingBaseURL + pathRefundStatus
}

func (e Environment) baseURL() string {
	if e == EnvProduction {
		return productionBaseURL
	}
	return stagingBaseURL
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client for API calls.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) {
		cl.httpClient = c
	}
}

// WithReadTimeout sets the overall HTTP request timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(cl *Client) {
		cl.httpClient.Timeout = d
	}
}

// WithConnectTimeout sets the TLS handshake timeout.
func WithConnectTimeout(d time.Duration) Option {
	return func(cl *Client) {
		transport, ok := cl.httpClient.Transport.(*http.Transport)
		if !ok || transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
			cl.httpClient.Transport = transport
		}
		transport.TLSHandshakeTimeout = d
	}
}

// WithClientID sets the optional merchant client ID sent in the request head.
func WithClientID(id string) Option {
	return func(cl *Client) {
		cl.clientID = id
	}
}

// WithCallbackURL sets the default callback URL for transaction responses.
func WithCallbackURL(url string) Option {
	return func(cl *Client) {
		cl.callbackURL = url
	}
}
