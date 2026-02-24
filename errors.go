package paytm

import "fmt"

// ErrorCode represents a specific category of SDK error.
type ErrorCode string

const (
	ErrSignatureValidationFailed ErrorCode = "SIGNATURE_VALIDATION_FAILED"
	ErrMissingMandatoryParams    ErrorCode = "MISSING_MANDATORY_PARAMETERS"
	ErrMerchantNotInitialized    ErrorCode = "MISSING_MERCHANT_PROPERTY"
	ErrJSONConversionFailed      ErrorCode = "JSON_CONVERSION_FAILED"
	ErrAPICallFailed             ErrorCode = "API_CALL_FAILED"
)

// Error represents an SDK error with a code, message, and optional raw response.
type Error struct {
	Code    ErrorCode
	Message string
	Raw     string // raw JSON response body, if available
}

func (e *Error) Error() string {
	if e.Raw != "" {
		return fmt.Sprintf("paytm: %s: %s (raw: %s)", e.Code, e.Message, e.Raw)
	}
	return fmt.Sprintf("paytm: %s: %s", e.Code, e.Message)
}

func newError(code ErrorCode, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func newErrorWithRaw(code ErrorCode, msg, raw string) *Error {
	return &Error{Code: code, Message: msg, Raw: raw}
}
