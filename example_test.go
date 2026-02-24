package paytm_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	paytm "github.com/KriaaCompany/paytm-go-sdk"
)

func ExampleNewClient() {
	// Create a client for the staging environment
	client := paytm.NewClient(
		"YOUR_MID",
		"YOUR_MERCHANT_KEY__", // must be 16, 24, or 32 bytes
		"WEBSTAGING",
		paytm.EnvStaging,
	)

	_ = client // use client to call API methods
}

func ExampleNewClient_withOptions() {
	// Create a client with custom HTTP settings
	client := paytm.NewClient(
		"YOUR_MID",
		"YOUR_MERCHANT_KEY__", // must be 16, 24, or 32 bytes
		"DEFAULT",
		paytm.EnvProduction,
		paytm.WithHTTPClient(&http.Client{
			Timeout: 60 * time.Second,
		}),
	)

	_ = client
}

func ExampleClient_InitiateTransaction() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.InitiateTransaction(context.Background(), paytm.InitiateTransactionRequest{
		OrderID:   "ORDER_001",
		TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
		UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
		ChannelID: "WEB",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("TxnToken:", resp.TxnToken)
	fmt.Println("Status:", resp.ResultInfo.ResultStatus)
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
}

func ExampleClient_InitiateRefund() {
	client := paytm.NewClient("YOUR_MID", "YOUR_MERCHANT_KEY__", "WEBSTAGING", paytm.EnvStaging)

	resp, err := client.InitiateRefund(context.Background(), paytm.RefundRequest{
		OrderID:      "ORDER_001",
		RefID:        "REFUND_001",
		TxnID:        "TXN_12345",
		TxnType:      "REFUND",
		RefundAmount: "50.00",
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
	fmt.Println("Amount:", resp.RefundAmount)
}
