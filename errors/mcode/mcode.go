package mcode

// Code is the error code interface.
type Code interface {
	Code() int
	Message() string
	Detail() any
}

// Predefined error codes.
var (
	CodeNil                      = localCode{-1, "", nil}
	CodeOK                       = localCode{0, "OK", nil}
	CodeUnknown                  = localCode{1, "Unknown", nil}
	CodeInvalidRequest           = localCode{1000, "Invalid Request", nil}
	CodeInvalidParameter         = localCode{1001, "Invalid Parameter", nil}
	CodeMissingParameter         = localCode{1002, "Missing Parameter", nil}
	CodeValidationFailed         = localCode{1003, "Validation Failed", nil}
	CodeNotFound                 = localCode{1004, "Not Found", nil}
	CodeNotAuthorized            = localCode{1005, "Not Authorized", nil}
	CodeForbidden                = localCode{1006, "Forbidden", nil}
	CodeInternalError            = localCode{2000, "Internal Error", nil}
	CodeDbOperationError         = localCode{2001, "Database Operation Error", nil}
	CodeInternalPanic            = localCode{2002, "Internal Panic", nil}
	CodeServerBusy               = localCode{2003, "Server Busy", nil}
	CodeRateLimitExceeded        = localCode{2004, "Rate Limit Exceeded", nil}
	CodeInvalidOperation         = localCode{3000, "Invalid Operation", nil}
	CodeInvalidConfiguration     = localCode{3001, "Invalid Configuration", nil}
	CodeMissingConfiguration     = localCode{3002, "Missing Configuration", nil}
	CodeNotImplemented           = localCode{3003, "Not Implemented", nil}
	CodeNotSupported             = localCode{3004, "Not Supported", nil}
	CodeOperationFailed          = localCode{3005, "Operation Failed", nil}
	CodeSecurityReason           = localCode{4000, "Security Reason", nil}
	CodeBusinessValidationFailed = localCode{5000, "Business Validation Failed", nil}
)

// New creates a new error code.
func New(code int, message string, detail any) Code {
	return localCode{
		code:    code,
		message: message,
		detail:  detail,
	}
}

// WithCode creates a new error code, using the given error code and detail.
func WithCode(code Code, detail any) Code {
	return localCode{
		code:    code.Code(),
		message: code.Message(),
		detail:  detail,
	}
}
