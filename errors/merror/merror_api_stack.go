package merror

import (
	"errors"
	"runtime"
)

// stack 表示程序计数器的堆栈
type stack []uintptr

const (
	// maxStackDepth 标记错误回溯的最大堆栈深度
	maxStackDepth = 64
)

// Cause 返回 `err` 的根本原因错误
func Cause(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(ICause); ok {
		return e.Cause()
	}
	if e, ok := err.(IUnwrap); ok {
		return Cause(e.Unwrap())
	}
	return err
}

// Stack 返回堆栈调用者信息的字符串
// 如果 `err` 不支持堆栈，则直接返回错误字符串
func Stack(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(IStack); ok {
		return e.Stack()
	}
	return err.Error()
}

// Current 创建并返回当前级别的错误
// 如果当前级别错误为 nil，则返回 nil
func Current(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(ICurrent); ok {
		return e.Current()
	}
	return err
}

// Unwrap 返回下一级错误
// 如果当前级别错误或下一级错误为 nil，则返回 nil
func Unwrap(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(IUnwrap); ok {
		return e.Unwrap()
	}
	return nil
}

// HasStack 检查并报告 `err` 是否实现了接口 `gerror.IStack`
func HasStack(err error) bool {
	_, ok := err.(IStack)
	return ok
}

// Equal 报告当前错误 `err` 是否等于错误 `target`
// 请注意，在 `Error` 的默认比较逻辑中，
// 如果两个错误的 `code` 和 `text` 都相同，则认为它们是相同的
func Equal(err, target error) bool {
	if err == target {
		return true
	}
	if e, ok := err.(IEqual); ok {
		return e.Equal(target)
	}
	if e, ok := target.(IEqual); ok {
		return e.Equal(err)
	}
	return false
}

// Is 报告当前错误 `err` 的链式错误中是否包含错误 `target`
// 有一个类似的函数 HasError，它是在 go 标准库的 errors.Is 之前设计和实现的
// 现在它是 go 标准库的 errors.Is 的别名，以保证与 go 标准库相同的性能
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 在 err 的错误链中查找与 target 匹配的第一个错误，如果找到，
// 则将 target 设置为该错误值并返回 true
//
// 错误链由 err 本身组成，后跟通过重复调用 Unwrap 获得的错误序列
//
// 如果错误的具体值可分配给 target 指向的值，或者错误具有方法 As(interface{}) bool
// 使得 As(target) 返回 true，则错误与 target 匹配。在后一种情况下，
// As 方法负责设置 target
//
// 如果 target 不是指向实现 error 的类型或任何接口类型的非 nil 指针，As 将会 panic
// 如果 err 为 nil，As 返回 false
func As(err error, target any) bool {
	return errors.As(err, target)
}

// callers 返回堆栈调用者
// 注意，这里只是检索调用者内存地址数组，而不是调用者信息
func callers(skip ...int) stack {
	var (
		pcs [maxStackDepth]uintptr
		n   = 3
	)
	if len(skip) > 0 {
		n += skip[0]
	}
	return pcs[:runtime.Callers(n, pcs[:])]
}
