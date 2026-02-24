# Paytm Payment Gateway Go SDK

Unofficial Go SDK for the [Paytm Payment Gateway](https://business.paytm.com/docs). Paytm officially supports Java, Node, PHP, and Python — this SDK fills the gap for Go.

**Zero external dependencies** — uses only the Go standard library.

## Installation

```bash
go get github.com/KriaaCompany/paytm-go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    paytm "github.com/KriaaCompany/paytm-go-sdk"
)

func main() {
    client := paytm.NewClient(
        "YOUR_MID",
        "YOUR_MERCHANT_KEY",
        "WEBSTAGING",
        paytm.EnvStaging,
    )

    resp, err := client.InitiateTransaction(context.Background(), paytm.InitiateTransactionRequest{
        OrderID:   "ORDER_001",
        TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
        UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
        ChannelID: "WEB",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("TxnToken:", resp.TxnToken)
}
```

## API Reference

### Creating a Client

```go
// Staging
client := paytm.NewClient("MID", "KEY", "WEBSTAGING", paytm.EnvStaging)

// Production
client := paytm.NewClient("MID", "KEY", "DEFAULT", paytm.EnvProduction)

// With custom HTTP client
client := paytm.NewClient("MID", "KEY", "DEFAULT", paytm.EnvProduction,
    paytm.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
)
```

### Initiate Transaction

Starts a new payment transaction and returns a transaction token for the Paytm checkout page.

```go
resp, err := client.InitiateTransaction(ctx, paytm.InitiateTransactionRequest{
    OrderID:   "ORDER_001",
    TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
    UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
    ChannelID: "WEB",
})
// resp.TxnToken — use this to open the Paytm payment page
```

### Get Payment Status

Check the status of a transaction.

```go
resp, err := client.GetPaymentStatus(ctx, paytm.PaymentStatusRequest{
    OrderID: "ORDER_001",
})
// resp.ResultInfo.ResultStatus — "TXN_SUCCESS", "TXN_FAILURE", "PENDING"
```

### Initiate Refund

Initiate a refund for a completed transaction.

```go
resp, err := client.InitiateRefund(ctx, paytm.RefundRequest{
    OrderID:      "ORDER_001",
    RefID:        "REFUND_001",
    TxnID:        "TXN_12345",
    TxnType:      "REFUND",
    RefundAmount: "50.00",
})
```

### Get Refund Status

Check the status of a refund.

```go
resp, err := client.GetRefundStatus(ctx, paytm.RefundStatusRequest{
    OrderID: "ORDER_001",
    RefID:   "REFUND_001",
})
```

## Error Handling

All API methods return `*paytm.Error` which includes a code and message:

```go
resp, err := client.InitiateTransaction(ctx, req)
if err != nil {
    var sdkErr *paytm.Error
    if errors.As(err, &sdkErr) {
        fmt.Println("Code:", sdkErr.Code)
        fmt.Println("Message:", sdkErr.Message)
        fmt.Println("Raw response:", sdkErr.Raw)
    }
}
```

Error codes:
- `SIGNATURE_VALIDATION_FAILED` — checksum generation or verification failed
- `MISSING_MANDATORY_PARAMETERS` — required request parameters are missing
- `MISSING_MERCHANT_PROPERTY` — merchant configuration is incomplete
- `JSON_CONVERSION_FAILED` — JSON marshaling/unmarshaling error
- `API_CALL_FAILED` — HTTP request failed or returned non-200

## Payment Modes

Filter payment modes using `PaymentModeFilter`:

```go
resp, err := client.InitiateTransaction(ctx, paytm.InitiateTransactionRequest{
    OrderID:   "ORDER_001",
    TxnAmount: paytm.Money{Value: "100.00", Currency: "INR"},
    UserInfo:  paytm.UserInfo{CustID: "CUST_001"},
    ChannelID: "WEB",
    PaymentModeFilter: &paytm.PaymentModeFilter{
        EnablePaymentModes: []paytm.PaymentMode{
            paytm.PaymentModeUPI,
            paytm.PaymentModeCreditCard,
        },
    },
})
```

Available modes: `PaymentModeBalance`, `PaymentModeUPI`, `PaymentModeCreditCard`, `PaymentModeDebitCard`, `PaymentModeNetBanking`, `PaymentModeEMI`, `PaymentModePPBL`.

## Environments

| Environment | Base URL |
|---|---|
| `paytm.EnvStaging` | `https://securegw-stage.paytm.in` |
| `paytm.EnvProduction` | `https://securegw.paytm.in` |

## License

MIT
