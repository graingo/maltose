package mcode

// Code 错误码接口
type Code interface {
	Code() int
	Message() string
	Detail() any
}

// 预定义错误码
var (
	// 基础错误码 (0-99)
	CodeNil     = localCode{-1, "", nil}
	CodeOK      = localCode{0, "成功", nil}
	CodeUnknown = localCode{1, "未知错误", nil}

	// 客户端错误 (100-199)
	CodeInvalidRequest   = localCode{100, "无效请求", nil}
	CodeInvalidParameter = localCode{101, "参数无效", nil}
	CodeMissingParameter = localCode{102, "参数缺失", nil}
	CodeValidationFailed = localCode{103, "数据验证失败", nil}
	CodeNotFound         = localCode{104, "资源不存在", nil}
	CodeNotAuthorized    = localCode{105, "未授权", nil}
	CodeForbidden        = localCode{106, "禁止访问", nil}

	// 服务端错误 (200-299)
	CodeInternalError    = localCode{200, "内部错误", nil}
	CodeDbOperationError = localCode{201, "数据库操作错误", nil}
	CodeInternalPanic    = localCode{202, "内部异常", nil}
	CodeServerBusy       = localCode{203, "服务器繁忙", nil}

	// 配置与操作错误 (300-399)
	CodeInvalidOperation     = localCode{300, "无效操作", nil}
	CodeInvalidConfiguration = localCode{301, "配置无效", nil}
	CodeMissingConfiguration = localCode{302, "配置缺失", nil}
	CodeNotImplemented       = localCode{303, "未实现", nil}
	CodeNotSupported         = localCode{304, "不支持", nil}
	CodeOperationFailed      = localCode{305, "操作失败", nil}

	// 安全相关错误 (400-499)
	CodeSecurityReason = localCode{400, "安全原因", nil}

	// 业务逻辑错误 (500-599)
	CodeBusinessValidationFailed = localCode{500, "业务验证失败", nil}
)

// New 创建一个新的错误码
func New(code int, message string, detail any) Code {
	return localCode{
		code:    code,
		message: message,
		detail:  detail,
	}
}

// WithCode 创建一个新的错误码，使用给定的错误码和详细信息
func WithCode(code Code, detail any) Code {
	return localCode{
		code:    code.Code(),
		message: code.Message(),
		detail:  detail,
	}
}
