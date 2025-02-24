package merror

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/mingzaily/maltose/errors/mcode"
)

// Error 错误结构体
type Error struct {
	error error
	code  mcode.Code
	stack []string
}

// New 创建一个新的基础错误
// 适用于创建一个全新的错误，默认使用 InternalError 错误码
// 示例：err := merror.New("数据库连接失败")
func New(text string, a ...any) error {
	return &Error{
		error: fmt.Errorf(text, a...),
		code:  mcode.InternalError,
		stack: callers(),
	}
}

// NewCode 创建一个带指定错误码的新错误
// 适用于创建一个全新的错误，并指定特定的错误码
// 示例：err := merror.NewCode(mcode.ValidationError, "用户名不能为空")
func NewCode(code mcode.Code, text string, a ...any) error {
	return &Error{
		error: fmt.Errorf(text, a...),
		code:  code,
		stack: callers(),
	}
}

// Wrap 包装已有错误，保持原错误码
// 适用于包装底层错误，添加上下文信息，同时保留原始错误的错误码
// 示例：return merror.Wrap(err, "查询用户信息失败")
func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: fmt.Errorf("%s: %w", text, err),
		code:  CodeFromError(err),
		stack: callers(),
	}
}

// WrapCode 包装已有错误，并指定新的错误码
// 适用于包装底层错误，但需要改变错误码的场景
// 示例：return merror.WrapCode(err, mcode.ValidationError, "用户验证失败")
func WrapCode(err error, code mcode.Code, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: fmt.Errorf("%s: %w", text, err),
		code:  code,
		stack: callers(),
	}
}

// WrapMsg 包装错误，保留原错误码但替换提示信息
// 适用于需要修改错误提示但保持原始错误码的场景
// 示例：return merror.WrapMsg(err, "用户注册失败，请稍后重试")
func WrapMsg(err error, text string, a ...any) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: fmt.Errorf(text, a...),
		code:  CodeFromError(err),
		stack: callers(),
	}
}

// Stack 获取错误堆栈信息
func (e *Error) Stack() string {
	if e == nil {
		return ""
	}
	return formatStack(e.stack)
}

func (e *Error) Error() string {
	return e.error.Error()
}

// Code 获取错误码
func (e *Error) Code() mcode.Code {
	return e.code
}

// CodeFromError 从错误中获取错误码
func CodeFromError(err error) mcode.Code {
	if err == nil {
		return mcode.Success
	}
	if e, ok := err.(*Error); ok {
		return e.code
	}
	return mcode.InternalError
}

// 获取调用堆栈
func callers() []string {
	var (
		pc     = make([]uintptr, 32)
		n      = runtime.Callers(2, pc)
		frames = runtime.CallersFrames(pc[:n])
		stacks = make([]string, 0)
	)

	for {
		frame, more := frames.Next()
		// 过滤掉 merror 包内的调用
		if !strings.Contains(frame.File, "maltose/errors/merror") {
			stacks = append(stacks, fmt.Sprintf("%s\n    %s:%d", frame.Function, frame.File, frame.Line))
		}
		if !more {
			break
		}
	}
	return stacks
}

// 格式化堆栈信息
func formatStack(stack []string) string {
	if len(stack) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	for i, s := range stack {
		buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, s))
	}
	return buffer.String()
}
