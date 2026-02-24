package paytm_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	paytm "github.com/KriaaCompany/paytm-go-sdk"
)

func ExampleNewClient() {
	client := paytm.NewClient(
		"YOUR_MID",
		"YOUR_MERCHANT_KEY__", // must be 16, 24, or 32 bytes
		"WEBSTAGING",
		paytm.EnvStaging,
	)
	_ = client
}

func ExampleNewClient_withOptions() {
	client := paytm.NewClient(
		"YOUR_MID",
		"YOUR_MERCHANT_KEY__",
		"DEFAULT",
		paytm.EnvProduction,
		paytm.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
		paytm.WithClientID("YOUR_CLIENT_ID"),
		paytm.WithCallbackURL("https://yoursite.com/paytm/callback"),
	)
	_ = client
}

func ExampleClient_InitiateTransaction() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.InitiateTransaction(context.Background(), paytm.InitiateTransactionRequest{
		ChannelID: paytm.ChannelWeb,
		OrderID:   "ORDER_001",
		TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
		UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("TxnToken:", resp.TxnToken)
	fmt.Println("Status:", resp.ResultInfo.ResultStatus)
}

func ExampleClient_InitiateTransaction_withPaymentModes() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.InitiateTransaction(context.Background(), paytm.InitiateTransactionRequest{
		ChannelID: paytm.ChannelWeb,
		OrderID:   "ORDER_002",
		TxnAmount: paytm.Money{Value: "500.00", Currency: "INR"},
		UserInfo: paytm.UserInfo{
			CustID: "CUST_001",
			Mobile: "9999999999",
			Email:  "user@example.com",
		},
		EnablePaymentMode: []paytm.PaymentMode{
			{Mode: paytm.PaymentModeNameUPI},
			{Mode: paytm.PaymentModeNameCreditCard, Channels: []string{"VISA", "MASTERCARD"}},
		},
		DisablePaymentMode: []paytm.PaymentMode{
			{Mode: paytm.PaymentModeNameEMI},
		},
		ExtendInfo: &paytm.ExtendInfo{
			UDF1:       "custom-value-1",
			MercUnqRef: "my-internal-ref-001",
		},
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("TxnToken:", resp.TxnToken)
}

func ExampleClient_GetPaymentStatus() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.GetPaymentStatus(context.Background(), paytm.PaymentStatusRequest{
		OrderID: "ORDER_001",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("TxnID:", resp.TxnID)
	fmt.Println("Status:", resp.ResultInfo.ResultStatus)
	fmt.Println("Amount:", resp.TxnAmount)
	fmt.Println("Mode:", resp.PaymentMode)
}

func ExampleClient_InitiateRefund() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.InitiateRefund(context.Background(), paytm.RefundRequest{
		OrderID:      "ORDER_001",
		RefID:        "REFUND_001",
		TxnID:        "TXN_12345",
		TxnType:      "REFUND",
		RefundAmount: "50.00",
		Comments:     "Customer requested refund",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("RefundID:", resp.RefundID)
	fmt.Println("Status:", resp.ResultInfo.ResultStatus)
}

func ExampleClient_GetRefundStatus() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.GetRefundStatus(context.Background(), paytm.RefundStatusRequest{
		OrderID: "ORDER_001",
		RefID:   "REFUND_001",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("RefundID:", resp.RefundID)
	fmt.Println("Status:", resp.ResultInfo.ResultStatus)
	fmt.Println("Total Refunded:", resp.TotalRefundedAmount)
}
