package paytm

// PaymentMode represents a payment method type.
type PaymentMode string

const (
	PaymentModeBalance    PaymentMode = "BALANCE"
	PaymentModeUPI        PaymentMode = "UPI"
	PaymentModeCreditCard PaymentMode = "CREDIT_CARD"
	PaymentModeDebitCard  PaymentMode = "DEBIT_CARD"
	PaymentModeNetBanking PaymentMode = "NET_BANKING"
	PaymentModeEMI        PaymentMode = "EMI"
	PaymentModePPBL       PaymentMode = "PPBL"
)

// Money represents a monetary amount with currency.
type Money struct {
	Value    string `json:"value"`
	Currency string `json:"currency,omitempty"`
}

// UserInfo holds customer information for a transaction.
type UserInfo struct {
	CustID    string `json:"custId"`
	Mobile    string `json:"mobile,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Address   string `json:"address,omitempty"`
	Pincode   string `json:"pincode,omitempty"`
}

// PaymentModeFilter controls which payment modes to enable or disable.
type PaymentModeFilter struct {
	EnablePaymentModes  []PaymentMode `json:"enablePaymentMode,omitempty"`
	DisablePaymentModes []PaymentMode `json:"disablePaymentMode,omitempty"`
}

// ResultInfo holds the result status from Paytm API responses.
type ResultInfo struct {
	ResultStatus string `json:"resultStatus"`
	ResultCode   string `json:"resultCode"`
	ResultMsg    string `json:"resultMsg"`
}

// --- Request Models ---

// InitiateTransactionRequest holds parameters for the initiateTransaction API.
type InitiateTransactionRequest struct {
	OrderID           string             `json:"orderId"`
	ChannelID         string             `json:"channelId,omitempty"`
	TxnAmount         Money              `json:"txnAmount"`
	UserInfo          UserInfo           `json:"userInfo"`
	PaymentModeFilter *PaymentModeFilter `json:"enablePaymentMode,omitempty"`
	PromoCode         string             `json:"promoCode,omitempty"`
	CallbackURL       string             `json:"callbackUrl,omitempty"`
}

// PaymentStatusRequest holds parameters for the getPaymentStatus API.
type PaymentStatusRequest struct {
	OrderID string `json:"orderId"`
}

// RefundRequest holds parameters for the refund API.
type RefundRequest struct {
	OrderID      string `json:"orderId"`
	RefID        string `json:"refId"`
	TxnID        string `json:"txnId"`
	TxnType      string `json:"txnType"`
	RefundAmount string `json:"refundAmount"`
}

// RefundStatusRequest holds parameters for the refundStatus API.
type RefundStatusRequest struct {
	OrderID string `json:"orderId"`
	RefID   string `json:"refId,omitempty"`
}

// --- Response Models ---

// InitiateTransactionResponse holds the response from the initiateTransaction API.
type InitiateTransactionResponse struct {
	ResultInfo  ResultInfo `json:"resultInfo"`
	TxnToken    string     `json:"txnToken"`
	IsPromoCode bool       `json:"isPromoCodeValid"`
	Authenticated bool     `json:"authenticated"`
}

// PaymentStatusResponse holds the response from the getPaymentStatus API.
type PaymentStatusResponse struct {
	ResultInfo  ResultInfo `json:"resultInfo"`
	TxnID       string     `json:"txnId"`
	BankTxnID   string     `json:"bankTxnId"`
	OrderID     string     `json:"orderId"`
	TxnAmount   string     `json:"txnAmount"`
	TxnDate     string     `json:"txnDate"`
	GatewayName string     `json:"gatewayName"`
	PaymentMode string     `json:"paymentMode"`
	BankName    string     `json:"bankName"`
	MID         string     `json:"mid"`
	RefundAmt   string     `json:"refundAmt"`
}

// RefundResponse holds the response from the refund API.
type RefundResponse struct {
	ResultInfo   ResultInfo `json:"resultInfo"`
	TxnID        string     `json:"txnId"`
	OrderID      string     `json:"orderId"`
	RefundID     string     `json:"refundId"`
	RefundAmount string     `json:"refundAmount"`
}

// RefundStatusResponse holds the response from the refundStatus API.
type RefundStatusResponse struct {
	ResultInfo   ResultInfo `json:"resultInfo"`
	TxnID        string     `json:"txnId"`
	OrderID      string     `json:"orderId"`
	RefundID     string     `json:"refundId"`
	RefundAmount string     `json:"refundAmount"`
}
