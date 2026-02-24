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

// mockPaytmServer creates an httptest server that mimics Paytm API responses.
// It verifies request signatures and returns signed responses.
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

		xReqID := r.Header.Get("X-Request-ID")
		if !strings.HasPrefix(xReqID, "GO-SDK:") {
			t.Errorf("expected X-Request-ID starting with GO-SDK:, got %s", xReqID)
		}

		// Verify mid and orderId query params
		mid := r.URL.Query().Get("mid")
		orderID := r.URL.Query().Get("orderId")
		if mid == "" {
			t.Error("missing mid query parameter")
		}
		if orderID == "" {
			t.Error("missing orderId query parameter")
		}

		// Parse the request envelope
		var reqEnvelope apiRequest
		if err := json.NewDecoder(r.Body).Decode(&reqEnvelope); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Verify the request signature
		valid, err := checksum.Verify(string(reqEnvelope.Body), reqEnvelope.Head.Signature, merchantKey)
		if err != nil {
			t.Errorf("failed to verify request signature: %v", err)
		}
		if !valid {
			t.Error("request signature verification failed")
		}

		// Determine response based on path
		var respBody interface{}
		path := r.URL.Path

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

		// Marshal response body and sign it
		bodyBytes, _ := json.Marshal(respBody)
		sig, _ := checksum.Generate(string(bodyBytes), merchantKey)

		resp := apiResponse{
			Body: bodyBytes,
		}
		resp.Head.Signature = sig

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	return &Client{
		mid:         "TEST_MID_123",
		merchantKey: testMerchantKey,
		website:     "WEBSTAGING",
		env:         EnvStaging,
		httpClient:  &http.Client{},
	}
}

// overrideBaseURL temporarily replaces the environment's base URL for testing.
// We achieve this by injecting a custom transport that rewrites the URL.
type rewriteTransport struct {
	base      http.RoundTripper
	targetURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the scheme+host with our test server
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
			Transport: &rewriteTransport{
				base:      http.DefaultTransport,
				targetURL: serverURL,
			},
		},
	}
}

func TestInitiateTransaction(t *testing.T) {
	server := mockPaytmServer(t, testMerchantKey)
	defer server.Close()

	client := newTestClientWithServer(t, server.URL)

	resp, err := client.InitiateTransaction(context.Background(), InitiateTransactionRequest{
		OrderID:   "ORDER_001",
		TxnAmount: Money{Value: "100.00", Currency: "INR"},
		UserInfo:  UserInfo{CustID: "CUST_001"},
		ChannelID: "WEB",
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
	if resp.ResultInfo.ResultCode != "0000" {
		t.Errorf("expected resultCode 0000, got %s", resp.ResultInfo.ResultCode)
	}
}

func TestInitiateTransaction_MissingParams(t *testing.T) {
	client := newTestClient(t, "")

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
				t.Errorf("expected error code %s, got %s", ErrMissingMandatoryParams, sdkErr.Code)
			}
			if !strings.Contains(sdkErr.Message, tt.want) {
				t.Errorf("expected error message to contain %q, got %q", tt.want, sdkErr.Message)
			}
		})
	}
}

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
	if resp.TxnAmount != "100.00" {
		t.Errorf("expected txnAmount 100.00, got %s", resp.TxnAmount)
	}
	if resp.ResultInfo.ResultStatus != "TXN_SUCCESS" {
		t.Errorf("expected resultStatus TXN_SUCCESS, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestGetPaymentStatus_MissingOrderID(t *testing.T) {
	client := newTestClient(t, "")

	_, err := client.GetPaymentStatus(context.Background(), PaymentStatusRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

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
		t.Errorf("expected resultStatus PENDING, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestInitiateRefund_MissingParams(t *testing.T) {
	client := newTestClient(t, "")

	tests := []struct {
		name string
		req  RefundRequest
		want string
	}{
		{
			name: "missing orderId",
			req:  RefundRequest{RefID: "R1", TxnID: "T1", RefundAmount: "10.00"},
			want: "orderId",
		},
		{
			name: "missing refId",
			req:  RefundRequest{OrderID: "O1", TxnID: "T1", RefundAmount: "10.00"},
			want: "refId",
		},
		{
			name: "missing txnId",
			req:  RefundRequest{OrderID: "O1", RefID: "R1", RefundAmount: "10.00"},
			want: "txnId",
		},
		{
			name: "missing refundAmount",
			req:  RefundRequest{OrderID: "O1", RefID: "R1", TxnID: "T1"},
			want: "refundAmount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.InitiateRefund(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

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
		t.Errorf("expected resultStatus TXN_SUCCESS, got %s", resp.ResultInfo.ResultStatus)
	}
}

func TestGetRefundStatus_MissingOrderID(t *testing.T) {
	client := newTestClient(t, "")

	_, err := client.GetRefundStatus(context.Background(), RefundStatusRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("MID", "KEY_1234567890AB", "WEBSTAGING", EnvStaging)

	if client.mid != "MID" {
		t.Errorf("expected mid MID, got %s", client.mid)
	}
	if client.merchantKey != "KEY_1234567890AB" {
		t.Errorf("expected merchantKey KEY_1234567890AB, got %s", client.merchantKey)
	}
	if client.website != "WEBSTAGING" {
		t.Errorf("expected website WEBSTAGING, got %s", client.website)
	}
	if client.env != EnvStaging {
		t.Errorf("expected env STAGE, got %s", client.env)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customHTTPClient := &http.Client{}
	client := NewClient("MID", "KEY", "WEB", EnvProduction, WithHTTPClient(customHTTPClient))

	if client.httpClient != customHTTPClient {
		t.Error("expected custom httpClient")
	}
	if client.env != EnvProduction {
		t.Error("expected production environment")
	}
}

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
		t.Errorf("expected error code %s, got %s", ErrAPICallFailed, sdkErr.Code)
	}
}
