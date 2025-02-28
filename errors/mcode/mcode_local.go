package mcode

import "fmt"

// localCode 是错误码的实现
type localCode struct {
	code    int
	message string
	detail  any
}

// Code 返回错误码
func (c localCode) Code() int {
	return c.code
}

// Message 返回错误信息
func (c localCode) Message() string {
	return c.message
}

// Detail 返回当前错误码的详细信息，主要用于错误码的扩展字段。
func (c localCode) Detail() any {
	return c.detail
}

// String 返回错误码的字符串表示
func (c localCode) String() string {
	if c.detail != nil {
		return fmt.Sprintf(`%d:%s %v`, c.code, c.message, c.detail)
	}
	if c.message != "" {
		return fmt.Sprintf(`%d:%s`, c.code, c.message)
	}
	return fmt.Sprintf(`%d`, c.code)
}
