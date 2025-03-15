package merror

import (
	"errors"
	"runtime"
)

// stack represents the stack of program counters.
type stack []uintptr

const (
	// maxStackDepth marks the maximum stack depth.
	maxStackDepth = 64
)

// Cause returns the root cause of `err`.
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

// Stack returns the string of the stack caller information.
// If `err` does not support stack, it will return the error string directly.
func Stack(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(IStack); ok {
		return e.Stack()
	}
	return err.Error()
}

// Current creates and returns the current level error.
// If the current level error is nil, it returns nil.
func Current(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(ICurrent); ok {
		return e.Current()
	}
	return err
}

// Unwrap returns the next level error.
// If the current level error or the next level error is nil, it returns nil.
func Unwrap(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(IUnwrap); ok {
		return e.Unwrap()
	}
	return nil
}

// HasStack checks and reports whether `err` implements the `gerror.IStack` interface.
func HasStack(err error) bool {
	_, ok := err.(IStack)
	return ok
}

// Equal reports whether `err` is equal to `target`.
// Note that in the default comparison logic of `Error`,
// if the `code` and `text` of the two errors are the same, it is considered that they are the same.
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

// Is reports whether `err` is in the chain of errors.
// There is a similar function `HasError`, it is designed and implemented before the `errors.Is` function of the go standard library.
// Now it is an alias of the `errors.Is` function of the go standard library, to ensure the same performance as the go standard library.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As searches for the first error in the chain of errors that matches `target`.
// If found, it sets `target` to the error value and returns true.
//
// The error chain consists of `err` itself, followed by the error sequence obtained by repeatedly calling `Unwrap`.
//
// If the specific value of the error can be assigned to the value pointed to by `target`, or the error has a method `As(interface{}) bool`
// so that `As(target)` returns true, then the error matches target. In the latter case,
// the As method is responsible for setting target.
//
// If target is not a pointer to a type that implements the error interface or any interface type, As will panic.
// If err is nil, As returns false.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// callers returns the stack caller.
// Note that this only retrieves the caller memory address array, not the caller information.
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
