package paytm

// ChannelID represents the payment channel (matches EChannelId in Node SDK).
type ChannelID string

const (
	ChannelApp    ChannelID = "APP"
	ChannelWeb    ChannelID = "WEB"
	ChannelWAP    ChannelID = "WAP"
	ChannelSystem ChannelID = "SYSTEM"
)

// PaymentModeName is the name of a payment method used in PaymentMode.Mode.
type PaymentModeName string

const (
	PaymentModeNameBalance    PaymentModeName = "BALANCE"
	PaymentModeNameUPI        PaymentModeName = "UPI"
	PaymentModeNameCreditCard PaymentModeName = "CREDIT_CARD"
	PaymentModeNameDebitCard  PaymentModeName = "DEBIT_CARD"
	PaymentModeNameNetBanking PaymentModeName = "NET_BANKING"
	PaymentModeNameEMI        PaymentModeName = "EMI"
	PaymentModeNamePPBL       PaymentModeName = "PPBL"
)

// PaymentMode represents a payment method with optional channel filter.
// Matches the Node SDK PaymentMode model (mode + channels[]).
type PaymentMode struct {
	Mode     PaymentModeName `json:"mode"`
	Channels []string        `json:"channels,omitempty"`
}

// UserSubWalletType represents sub-wallet categories for payment filtering.
type UserSubWalletType string

const (
	SubWalletFood                  UserSubWalletType = "FOOD"
	SubWalletGift                  UserSubWalletType = "GIFT"
	SubWalletMultiPurposeGift      UserSubWalletType = "MULTI_PURPOSE_GIFT"
	SubWalletToll                  UserSubWalletType = "TOLL"
	SubWalletClosedLoop            UserSubWalletType = "CLOSED_LOOP_WALLET"
	SubWalletClosedLoopSub         UserSubWalletType = "CLOSED_LOOP_SUB_WALLET"
	SubWalletFuel                  UserSubWalletType = "FUEL"
	SubWalletInternationalTransfer UserSubWalletType = "INTERNATIONAL_FUNDS_TRANSFER"
	SubWalletCashback              UserSubWalletType = "CASHBACK"
	SubWalletGiftVoucher           UserSubWalletType = "GIFT_VOUCHER"
	SubWalletCommunication         UserSubWalletType = "COMMUNICATION"
)

// Money represents a monetary amount with currency.
type Money struct {
	Value    string `json:"value"`
	Currency string `json:"currency,omitempty"`
}

// UserInfo holds customer information. CustID is required.
type UserInfo struct {
	CustID    string `json:"custId"`
	Mobile    string `json:"mobile,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Address   string `json:"address,omitempty"`
	Pincode   string `json:"pincode,omitempty"`
}

// ExtendInfo holds optional extended / user-defined fields for a transaction.
type ExtendInfo struct {
	UDF1               string                 `json:"udf1,omitempty"`
	UDF2               string                 `json:"udf2,omitempty"`
	UDF3               string                 `json:"udf3,omitempty"`
	MercUnqRef         string                 `json:"mercUnqRef,omitempty"`
	Comments           string                 `json:"comments,omitempty"`
	AmountToBeRefunded string                 `json:"amountToBeRefunded,omitempty"`
	SubwalletAmount    map[string]interface{} `json:"subwalletAmount,omitempty"`
}

// GoodsInfo holds product-level details for an order.
type GoodsInfo struct {
	MerchantGoodsID    string      `json:"merchantGoodsId,omitempty"`
	MerchantShippingID string      `json:"merchantShippingId,omitempty"`
	SnapshotURL        string      `json:"snapshotUrl,omitempty"`
	Description        string      `json:"description,omitempty"`
	Category           string      `json:"category,omitempty"`
	Quantity           string      `json:"quantity,omitempty"`
	Unit               string      `json:"unit,omitempty"`
	Price              *Money      `json:"price,omitempty"`
	ExtendInfo         *ExtendInfo `json:"extendInfo,omitempty"`
}

// ShippingInfo holds shipping/delivery details for an order.
type ShippingInfo struct {
	MerchantShippingID string `json:"merchantShippingId,omitempty"`
	TrackingNo         string `json:"trackingNo,omitempty"`
	Carrier            string `json:"carrier,omitempty"`
	ChargeAmount       *Money `json:"chargeAmount,omitempty"`
	CountryName        string `json:"countryName,omitempty"`
	StateName          string `json:"stateName,omitempty"`
	CityName           string `json:"cityName,omitempty"`
	Address1           string `json:"address1,omitempty"`
	Address2           string `json:"address2,omitempty"`
	FirstName          string `json:"firstName,omitempty"`
	LastName           string `json:"lastName,omitempty"`
	MobileNo           string `json:"mobileNo,omitempty"`
	ZipCode            string `json:"zipCode,omitempty"`
	Email              string `json:"email,omitempty"`
}

// ChildTransaction holds details of a child/split payment within a transaction.
type ChildTransaction struct {
	TxnID        string `json:"txnId,omitempty"`
	PaymentMode  string `json:"paymentMode,omitempty"`
	TxnAmount    string `json:"txnAmount,omitempty"`
	Gateway      string `json:"gateway,omitempty"`
	BankTxnID    string `json:"bankTxnId,omitempty"`
	BankName     string `json:"bankName,omitempty"`
	Status       string `json:"status,omitempty"`
	CardIndexNo  string `json:"cardIndexNo,omitempty"`
	MaskedCardNo string `json:"maskedCardNo,omitempty"`
}

// ResultInfo holds the result status returned by all Paytm API responses.
type ResultInfo struct {
	ResultStatus string `json:"resultStatus"`
	ResultCode   string `json:"resultCode"`
	ResultMsg    string `json:"resultMsg"`
	IsRedirect   bool   `json:"isRedirect,omitempty"`
}

// --- Request Models ---

// InitiateTransactionRequest holds all parameters for the initiateTransaction API.
// ChannelID is placed in the request head; all other fields go in the body.
type InitiateTransactionRequest struct {
	// ChannelID is sent in the request head (APP, WEB, WAP, SYSTEM).
	ChannelID ChannelID `json:"-"`

	OrderID                string        `json:"orderId"`
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

// PaymentStatusRequest holds parameters for the getPaymentStatus API.
type PaymentStatusRequest struct {
	OrderID string `json:"orderId"`
	TxnType string `json:"txnType,omitempty"`
}

// RefundRequest holds parameters for the async refund API.
type RefundRequest struct {
	OrderID              string        `json:"orderId"`
	RefID                string        `json:"refId"`
	TxnID                string        `json:"txnId"`
	TxnType              string        `json:"txnType"`
	RefundAmount         string        `json:"refundAmount"`
	Comments             string        `json:"comments,omitempty"`
	PreferredDestination string        `json:"preferredDestination,omitempty"`
	// RequestID defaults to RefID if empty (matches Node SDK getRequestId()).
	RequestID            string        `json:"requestId,omitempty"`
	SubwalletAmount      []interface{} `json:"subwalletAmount,omitempty"`
	ExtraParamsMap       map[string]interface{} `json:"extraParamsMap,omitempty"`
}

// RefundStatusRequest holds parameters for the refundStatus API.
// Both OrderID and RefID are required.
type RefundStatusRequest struct {
	OrderID string `json:"orderId"`
	RefID   string `json:"refId"`
}

// --- Response Models ---

// InitiateTransactionResponse is the response from the initiateTransaction API.
type InitiateTransactionResponse struct {
	ResultInfo     ResultInfo `json:"resultInfo"`
	TxnToken       string     `json:"txnToken"`
	IsPromoCode    bool       `json:"isPromoCodeValid"`
	Authenticated  bool       `json:"authenticated"`
	SubscriptionID string     `json:"subscriptionId,omitempty"`
	CallbackURL    string     `json:"callbackUrl,omitempty"`
}

// PaymentStatusResponse is the response from the getPaymentStatus API.
type PaymentStatusResponse struct {
	ResultInfo                  ResultInfo         `json:"resultInfo"`
	TxnID                       string             `json:"txnId"`
	BankTxnID                   string             `json:"bankTxnId"`
	OrderID                     string             `json:"orderId"`
	TxnAmount                   string             `json:"txnAmount"`
	TxnType                     string             `json:"txnType"`
	TxnDate                     string             `json:"txnDate"`
	GatewayName                 string             `json:"gatewayName"`
	PaymentMode                 string             `json:"paymentMode"`
	BankName                    string             `json:"bankName"`
	MID                         string             `json:"mid"`
	RefundAmt                   string             `json:"refundAmt"`
	RefundID                    string             `json:"refundId"`
	RefID                       string             `json:"refId"`
	ChildTransactions           []ChildTransaction `json:"childTransaction,omitempty"`
	SubsID                      string             `json:"subsId,omitempty"`
	MerchantUniqueReference     string             `json:"merchantUniqueReference,omitempty"`
	BlockedAmount               string             `json:"blockedAmount,omitempty"`
	PreAuthID                   string             `json:"preAuthId,omitempty"`
	CustomMerchantResponse      string             `json:"customMerchantResponse,omitempty"`
	CustomChecksumString        string             `json:"customChecksumString,omitempty"`
	MaskedCardNo                string             `json:"maskedCardNo,omitempty"`
	CardIndexNo                 string             `json:"cardIndexNo,omitempty"`
	MaskedCustomerMobileNumber  string             `json:"maskedCustomerMobileNumber,omitempty"`
	PosID                       string             `json:"posId,omitempty"`
	UniqueReferenceLabel        string             `json:"uniqueReferenceLabel,omitempty"`
	UniqueReferenceValue        string             `json:"uniqueReferenceValue,omitempty"`
	PCCCode                     string             `json:"pccCode,omitempty"`
	PRN                         string             `json:"prn,omitempty"`
	UDF1                        string             `json:"udf1,omitempty"`
	UDF2                        string             `json:"udf2,omitempty"`
	UDF3                        string             `json:"udf3,omitempty"`
	Comments                    string             `json:"comments,omitempty"`
	CurrentTxnCount             string             `json:"currentTxnCount,omitempty"`
	LoyaltyPoints               string             `json:"loyaltyPoints,omitempty"`
}

// RefundResponse is the response from the async refund API.
type RefundResponse struct {
	ResultInfo   ResultInfo `json:"resultInfo"`
	TxnID        string     `json:"txnId"`
	OrderID      string     `json:"orderId"`
	MID          string     `json:"mid"`
	TxnAmount    string     `json:"txnAmount"`
	RefundID     string     `json:"refundId"`
	RefundAmount string     `json:"refundAmount"`
	RefID        string     `json:"refId"`
}

// RefundStatusResponse is the response from the refundStatus API.
type RefundStatusResponse struct {
	ResultInfo          ResultInfo `json:"resultInfo"`
	TxnID               string     `json:"txnId"`
	OrderID             string     `json:"orderId"`
	MID                 string     `json:"mid"`
	TxnAmount           string     `json:"txnAmount"`
	RefundAmount        string     `json:"refundAmount"`
	TxnDate             string     `json:"txnDate"`
	TotalRefundedAmount string     `json:"totalRefundedAmount"`
	RefundDate          string     `json:"refundDate"`
	RefID               string     `json:"refId"`
	BankTxnID           string     `json:"bankTxnId"`
	TxnType             string     `json:"txnType"`
	GatewayName         string     `json:"gatewayName"`
	BankName            string     `json:"bankName"`
	PaymentMode         string     `json:"paymentMode"`
	RefundID            string     `json:"refundId"`
	RefundType          string     `json:"refundType"`
	SSOID               string     `json:"ssoId"`
}
