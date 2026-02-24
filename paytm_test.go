package paytm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KriaaCompany/paytm-go-sdk/internal/checksum"
)

const testMerchantKey = "kbzk1DSbJiV_O3p5" // 16-byte test key

// mockPaytmServer creates an httptest.Server that mimics Paytm API responses.
// It verifies incoming request signatures and returns signed responses.
func mockPaytmServer(t *testing.T, merchantKey string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if xid := r.Header.Get("X-Request-ID"); !strings.HasPrefix(xid, "GO-SDK:") {
			t.Errorf("expected X-Request-ID starting with GO-SDK:, got %s", xid)
		}

		// Parse the request envelope
		var reqEnvelope apiRequest
		if err := json.NewDecoder(r.Body).Decode(&reqEnvelope); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Verify head version
		if reqEnvelope.Head.Version != headVersion {
			t.Errorf("expected head version %s, got %s", headVersion, reqEnvelope.Head.Version)
		}

		// Verify request timestamp is a non-empty epoch ms string
		if reqEnvelope.Head.RequestTimestamp == "" {
			t.Error("expected non-empty requestTimestamp")
		}

		// Verify request signature
		valid, err := checksum.Verify(string(reqEnvelope.Body), reqEnvelope.Head.Signature, merchantKey)
		if err != nil {
			t.Errorf("failed to verify request signature: %v", err)
		}
		if !valid {
			t.Error("request signature verification failed")
		}

		path := r.URL.Path

		// Validate endpoint-specific rules
		switch {
		case strings.Contains(path, "initiateTransaction"):
			// Must have query params mid and orderId
			if r.URL.Query().Get("mid") == "" {
				t.Error("initiateTransaction: missing mid query param")
			}
			if r.URL.Query().Get("orderId") == "" {
				t.Error("initiateTransaction: missing orderId query param")
			}
			// Verify requestType in body
			var bodyCheck struct {
				RequestType string `json:"requestType"`
			}
			json.Unmarshal(reqEnvelope.Body, &bodyCheck)
			if bodyCheck.RequestType != "Payment" {
				t.Errorf("expected requestType Payment, got %q", bodyCheck.RequestType)
			}
		default:
			// Other endpoints must NOT have query params
			if r.URL.Query().Get("mid") != "" || r.URL.Query().Get("orderId") != "" {
				t.Errorf("%s: must not have mid/orderId query params", path)
			}
		}

		// Build response body based on path
		var respBody interface{}
		orderID := r.URL.Query().Get("orderId")
		if orderID == "" {
			var bodyObj struct {
				OrderID string `json:"orderId"`
			}
			json.Unmarshal(reqEnvelope.Body, &bodyObj)
			orderID = bodyObj.OrderID
		}

		switch {
		case strings.Contains(path, "initiateTransaction"):
			respBody = InitiateTransactionResponse{
				ResultInfo: ResultInfo{
					ResultStatus: "S",
					ResultCode:   "0000",
					ResultMsg:    "Success",
				},
				TxnToken: "test_txn_token_12345",
			}
		case strings.Contains(path, "getPaymentStatus"):
			respBody = PaymentStatusResponse{
				ResultInfo: ResultInfo{
					ResultStatus: "TXN_SUCCESS",
					ResultCode:   "01",
					ResultMsg:    "Txn Success",
				},
				TxnID:       "20210101111212345678",
				BankTxnID:   "BANK123456",
				OrderID:     orderID,
				TxnAmount:   "100.00",
				PaymentMode: "UPI",
			}
		case strings.Contains(path, "async/refund"):
			respBody = RefundResponse{
				ResultInfo: ResultInfo{
					ResultStatus: "PENDING",
					ResultCode:   "601",
					ResultMsg:    "Refund request was raised for this transaction. But it is pending state",
				},
				TxnID:        "20210101111212345678",
				OrderID:      orderID,
				RefundID:     "REFUND_001",
				RefundAmount: "50.00",
			}
		case strings.Contains(path, "refundStatus"):
			respBody = RefundStatusResponse{
				ResultInfo: ResultInfo{
					ResultStatus: "TXN_SUCCESS",
					ResultCode:   "10",
					ResultMsg:    "Refund Successful",
				},
				TxnID:        "20210101111212345678",
				OrderID:      orderID,
				RefundID:     "REFUND_001",
				RefundAmount: "50.00",
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		bodyBytes, _ := json.Marshal(respBody)
		sig, _ := checksum.Generate(string(bodyBytes), merchantKey)

		resp := apiResponse{Body: bodyBytes}
		resp.Head.Signature = sig

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// rewriteTransport redirects all requests to a test server, preserving path and query.
type rewriteTransport struct {
	base      http.RoundTripper
	targetURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.targetURL, "http://")
	return t.base.RoundTrip(req)
}

func newTestClientWithServer(t *testing.T, serverURL string) *Client {
	t.Helper()
	return &Client{
		mid:         "TEST_MID_123",
		merchantKey: testMerchantKey,
		website:     "WEBSTAGING",
		env:         EnvStaging,
		httpClient: &http.Client{
			Transport: &rewriteTransport{base: http.DefaultTransport, targetURL: serverURL},
		},
	}
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	return &Client{
		mid:         "TEST_MID_123",
		merchantKey: testMerchantKey,
		website:     "WEBSTAGING",
		env:         EnvStaging,
		httpClient:  &http.Client{},
	}
}

// --- InitiateTransaction ---

func TestInitiateTransaction(t *testing.T) {
	server := mockPaytmServer(t, testMerchantKey)
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	resp, err := client.InitiateTransaction(context.Background(), InitiateTransactionRequest{
		ChannelID: ChannelWeb,
		OrderID:   "ORDER_001",
		TxnAmount: Money{Value: "100.00", Currency: "INR"},
		UserInfo:  UserInfo{CustID: "CUST_001"},
	})
	if err != nil {
		t.Fatalf("InitiateTransaction failed: %v", err)
	}
	if resp.TxnToken != "test_txn_token_12345" {
		t.Errorf("expected txnToken test_txn_token_12345, got %s", resp.TxnToken)
	}
	if resp.ResultInfo.ResultStatus != "S" {
		t.Errorf("expected resultStatus S, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestInitiateTransaction_WithPaymentModes(t *testing.T) {
	server := mockPaytmServer(t, testMerchantKey)
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	resp, err := client.InitiateTransaction(context.Background(), InitiateTransactionRequest{
		ChannelID: ChannelWeb,
		OrderID:   "ORDER_002",
		TxnAmount: Money{Value: "200.00", Currency: "INR"},
		UserInfo:  UserInfo{CustID: "CUST_002"},
		EnablePaymentMode: []PaymentMode{
			{Mode: PaymentModeNameUPI},
			{Mode: PaymentModeNameCreditCard, Channels: []string{"VISA"}},
		},
		DisablePaymentMode: []PaymentMode{
			{Mode: PaymentModeNameEMI},
		},
	})
	if err != nil {
		t.Fatalf("InitiateTransaction with payment modes failed: %v", err)
	}
	if resp.TxnToken == "" {
		t.Error("expected non-empty txnToken")
	}
}

func TestInitiateTransaction_MissingParams(t *testing.T) {
	client := newTestClient(t)

	tests := []struct {
		name string
		req  InitiateTransactionRequest
		want string
	}{
		{
			name: "missing orderId",
			req:  InitiateTransactionRequest{TxnAmount: Money{Value: "100.00"}, UserInfo: UserInfo{CustID: "C1"}},
			want: "orderId",
		},
		{
			name: "missing txnAmount",
			req:  InitiateTransactionRequest{OrderID: "O1", UserInfo: UserInfo{CustID: "C1"}},
			want: "txnAmount",
		},
		{
			name: "missing custId",
			req:  InitiateTransactionRequest{OrderID: "O1", TxnAmount: Money{Value: "100.00"}},
			want: "custId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.InitiateTransaction(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			sdkErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("expected *Error, got %T", err)
			}
			if sdkErr.Code != ErrMissingMandatoryParams {
				t.Errorf("expected %s, got %s", ErrMissingMandatoryParams, sdkErr.Code)
			}
			if !strings.Contains(sdkErr.Message, tt.want) {
				t.Errorf("expected message to contain %q, got %q", tt.want, sdkErr.Message)
			}
		})
	}
}

// --- GetPaymentStatus ---

func TestGetPaymentStatus(t *testing.T) {
	server := mockPaytmServer(t, testMerchantKey)
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	resp, err := client.GetPaymentStatus(context.Background(), PaymentStatusRequest{
		OrderID: "ORDER_001",
	})
	if err != nil {
		t.Fatalf("GetPaymentStatus failed: %v", err)
	}
	if resp.OrderID != "ORDER_001" {
		t.Errorf("expected orderId ORDER_001, got %s", resp.OrderID)
	}
	if resp.ResultInfo.ResultStatus != "TXN_SUCCESS" {
		t.Errorf("expected TXN_SUCCESS, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestGetPaymentStatus_MissingOrderID(t *testing.T) {
	_, err := newTestClient(t).GetPaymentStatus(context.Background(), PaymentStatusRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- InitiateRefund ---

func TestInitiateRefund(t *testing.T) {
	server := mockPaytmServer(t, testMerchantKey)
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	resp, err := client.InitiateRefund(context.Background(), RefundRequest{
		OrderID:      "ORDER_001",
		RefID:        "REF_001",
		TxnID:        "TXN_001",
		TxnType:      "REFUND",
		RefundAmount: "50.00",
	})
	if err != nil {
		t.Fatalf("InitiateRefund failed: %v", err)
	}
	if resp.RefundAmount != "50.00" {
		t.Errorf("expected refundAmount 50.00, got %s", resp.RefundAmount)
	}
	if resp.ResultInfo.ResultStatus != "PENDING" {
		t.Errorf("expected PENDING, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestInitiateRefund_RequestIDDefaultsToRefID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var envelope apiRequest
		json.NewDecoder(r.Body).Decode(&envelope)

		var body struct {
			RequestID string `json:"requestId"`
			RefID     string `json:"refId"`
		}
		json.Unmarshal(envelope.Body, &body)

		if body.RequestID != body.RefID {
			t.Errorf("expected requestId to default to refId %q, got %q", body.RefID, body.RequestID)
		}

		// Return a minimal response
		respBody, _ := json.Marshal(RefundResponse{
			ResultInfo: ResultInfo{ResultStatus: "PENDING"},
		})
		sig, _ := checksum.Generate(string(respBody), testMerchantKey)
		resp := apiResponse{Body: respBody}
		resp.Head.Signature = sig
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)
	client.InitiateRefund(context.Background(), RefundRequest{
		OrderID:      "O1",
		RefID:        "REF_XYZ",
		TxnID:        "TXN_001",
		TxnType:      "REFUND",
		RefundAmount: "10.00",
		// RequestID intentionally empty — should default to RefID
	})
}

func TestInitiateRefund_MissingParams(t *testing.T) {
	client := newTestClient(t)

	tests := []struct {
		name string
		req  RefundRequest
	}{
		{"missing orderId", RefundRequest{RefID: "R1", TxnID: "T1", TxnType: "REFUND", RefundAmount: "10.00"}},
		{"missing refId", RefundRequest{OrderID: "O1", TxnID: "T1", TxnType: "REFUND", RefundAmount: "10.00"}},
		{"missing txnId", RefundRequest{OrderID: "O1", RefID: "R1", TxnType: "REFUND", RefundAmount: "10.00"}},
		{"missing txnType", RefundRequest{OrderID: "O1", RefID: "R1", TxnID: "T1", RefundAmount: "10.00"}},
		{"missing refundAmount", RefundRequest{OrderID: "O1", RefID: "R1", TxnID: "T1", TxnType: "REFUND"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.InitiateRefund(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			sdkErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("expected *Error, got %T", err)
			}
			if sdkErr.Code != ErrMissingMandatoryParams {
				t.Errorf("expected %s, got %s", ErrMissingMandatoryParams, sdkErr.Code)
			}
		})
	}
}

// --- GetRefundStatus ---

func TestGetRefundStatus(t *testing.T) {
	server := mockPaytmServer(t, testMerchantKey)
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	resp, err := client.GetRefundStatus(context.Background(), RefundStatusRequest{
		OrderID: "ORDER_001",
		RefID:   "REF_001",
	})
	if err != nil {
		t.Fatalf("GetRefundStatus failed: %v", err)
	}
	if resp.RefundAmount != "50.00" {
		t.Errorf("expected refundAmount 50.00, got %s", resp.RefundAmount)
	}
	if resp.ResultInfo.ResultStatus != "TXN_SUCCESS" {
		t.Errorf("expected TXN_SUCCESS, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestGetRefundStatus_MissingParams(t *testing.T) {
	client := newTestClient(t)

	tests := []struct {
		name string
		req  RefundStatusRequest
	}{
		{"missing orderId", RefundStatusRequest{RefID: "R1"}},
		{"missing refId", RefundStatusRequest{OrderID: "O1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetRefundStatus(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// --- Client construction ---

func TestNewClient(t *testing.T) {
	client := NewClient("MID", "KEY_1234567890AB", "WEBSTAGING", EnvStaging)

	if client.mid != "MID" {
		t.Errorf("expected mid MID, got %s", client.mid)
	}
	if client.merchantKey != "KEY_1234567890AB" {
		t.Errorf("unexpected merchantKey")
	}
	if client.website != "WEBSTAGING" {
		t.Errorf("expected website WEBSTAGING, got %s", client.website)
	}
	if client.env != EnvStaging {
		t.Errorf("expected EnvStaging, got %s", client.env)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customHTTPClient := &http.Client{}
	client := NewClient("MID", "KEY", "WEB", EnvProduction,
		WithHTTPClient(customHTTPClient),
		WithClientID("MY_CLIENT_ID"),
		WithCallbackURL("https://example.com/callback"),
	)

	if client.httpClient != customHTTPClient {
		t.Error("expected custom httpClient")
	}
	if client.env != EnvProduction {
		t.Error("expected production environment")
	}
	if client.clientID != "MY_CLIENT_ID" {
		t.Errorf("expected clientID MY_CLIENT_ID, got %s", client.clientID)
	}
	if client.callbackURL != "https://example.com/callback" {
		t.Errorf("unexpected callbackURL: %s", client.callbackURL)
	}
}

// --- URL construction ---

func TestEnvironmentURLs(t *testing.T) {
	if EnvStaging.initiateTransactionURL("MID", "ORD") !=
		"https://securestage.paytmpayments.com/theia/api/v1/initiateTransaction?mid=MID&orderId=ORD" {
		t.Error("staging initiateTransaction URL mismatch")
	}
	if EnvProduction.initiateTransactionURL("MID", "ORD") !=
		"https://secure.paytmpayments.com/theia/api/v1/initiateTransaction?mid=MID&orderId=ORD" {
		t.Error("production initiateTransaction URL mismatch")
	}
	if EnvStaging.paymentStatusURL() != "https://securestage.paytmpayments.com/merchant-status/api/v1/getPaymentStatus" {
		t.Error("staging paymentStatus URL mismatch")
	}
	if EnvProduction.paymentStatusURL() != "https://secure.paytmpayments.com/merchant-status/api/v1/getPaymentStatus" {
		t.Error("production paymentStatus URL mismatch")
	}
	// Production refundStatus uses a separate host
	if EnvProduction.refundStatusURL() != "https://pgp-ite.paytm.in/refund/api/v1/refundStatus" {
		t.Errorf("production refundStatus URL mismatch: %s", EnvProduction.refundStatusURL())
	}
	if EnvStaging.refundStatusURL() != "https://securestage.paytmpayments.com/refund/api/v1/refundStatus" {
		t.Error("staging refundStatus URL mismatch")
	}
}

// --- Error handling ---

func TestServerErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	_, err := client.GetPaymentStatus(context.Background(), PaymentStatusRequest{OrderID: "ORDER_001"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	sdkErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if sdkErr.Code != ErrAPICallFailed {
		t.Errorf("expected %s, got %s", ErrAPICallFailed, sdkErr.Code)
	}
}
