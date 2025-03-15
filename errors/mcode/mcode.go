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
	CodeInvalidRequest           = localCode{100, "Invalid Request", nil}
	CodeInvalidParameter         = localCode{101, "Invalid Parameter", nil}
	CodeMissingParameter         = localCode{102, "Missing Parameter", nil}
	CodeValidationFailed         = localCode{103, "Validation Failed", nil}
	CodeNotFound                 = localCode{104, "Not Found", nil}
	CodeNotAuthorized            = localCode{105, "Not Authorized", nil}
	CodeForbidden                = localCode{106, "Forbidden", nil}
	CodeInternalError            = localCode{200, "Internal Error", nil}
	CodeDbOperationError         = localCode{201, "Database Operation Error", nil}
	CodeInternalPanic            = localCode{202, "Internal Panic", nil}
	CodeServerBusy               = localCode{203, "Server Busy", nil}
	CodeInvalidOperation         = localCode{300, "Invalid Operation", nil}
	CodeInvalidConfiguration     = localCode{301, "Invalid Configuration", nil}
	CodeMissingConfiguration     = localCode{302, "Missing Configuration", nil}
	CodeNotImplemented           = localCode{303, "Not Implemented", nil}
	CodeNotSupported             = localCode{304, "Not Supported", nil}
	CodeOperationFailed          = localCode{305, "Operation Failed", nil}
	CodeSecurityReason           = localCode{400, "Security Reason", nil}
	CodeBusinessValidationFailed = localCode{500, "Business Validation Failed", nil}
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
