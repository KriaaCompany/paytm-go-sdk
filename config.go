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
	EnvProduction Environment = "PRODUCTION"
)

const (
	stagingBaseURL    = "https://securegw-stage.paytm.in"
	productionBaseURL = "https://securegw.paytm.in"

	sdkVersion = "1.0.0"
)

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

// WithConnectTimeout sets the TLS handshake and dial timeout.
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
