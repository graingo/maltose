package mcode

// Code 错误码接口
type Code interface {
	Code() int
	Message() string
}

type code struct {
	code    int
	message string
}

func (c code) Code() int {
	return c.code
}

func (c code) Message() string {
	return c.message
}

func New(c int, m string) Code {
	return &code{
		code:    c,
		message: m,
	}
}

// 预定义错误码
var (
	Success          = New(0, "success")
	InternalError    = New(500, "internal error")
	ValidationError  = New(400, "validation error")
	Unauthorized     = New(401, "unauthorized")
	Forbidden        = New(403, "forbidden")
	NotFound         = New(404, "not found")
	MethodNotAllowed = New(405, "method not allowed")
)
