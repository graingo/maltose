package merror_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/stretchr/testify/assert"
)

// Test for New, Newf, Wrap, Wrapf
func TestApi(t *testing.T) {
	// New
	err1 := merror.New("new error")
	assert.NotNil(t, err1)
	assert.Equal(t, "new error", err1.Error())
	assert.Equal(t, mcode.CodeNil, merror.Code(err1))

	// Newf
	err2 := merror.Newf("new error %d", 1)
	assert.NotNil(t, err2)
	assert.Equal(t, "new error 1", err2.Error())

	// Wrap
	err3 := merror.Wrap(errors.New("original"), "wrapped")
	assert.NotNil(t, err3)
	assert.Equal(t, "wrapped: original", err3.Error())

	// Wrapf
	err4 := merror.Wrapf(errors.New("original"), "wrapped %d", 1)
	assert.NotNil(t, err4)
	assert.Equal(t, "wrapped 1: original", err4.Error())

	// Wrap nil
	assert.Nil(t, merror.Wrap(nil, "wrap nil"))
	assert.Nil(t, merror.Wrapf(nil, "wrap nil %d", 1))
}

// Test for NewCode, NewCodef, WrapCode, WrapCodef, Code
func TestCode(t *testing.T) {
	// NewCode
	err1 := merror.NewCode(mcode.CodeValidationFailed, "validation failed")
	assert.NotNil(t, err1)
	assert.Equal(t, "validation failed", err1.Error())
	assert.Equal(t, mcode.CodeValidationFailed, merror.Code(err1))

	// NewCodef
	err2 := merror.NewCodef(mcode.CodeInvalidParameter, "param %s invalid", "name")
	assert.NotNil(t, err2)
	assert.Equal(t, "param name invalid", err2.Error())
	assert.Equal(t, mcode.CodeInvalidParameter, merror.Code(err2))

	// WrapCode
	err3 := merror.WrapCode(errors.New("original"), mcode.CodeDbOperationError, "db error")
	assert.NotNil(t, err3)
	assert.Equal(t, "db error: original", err3.Error())
	assert.Equal(t, mcode.CodeDbOperationError, merror.Code(err3))

	// WrapCodef
	err4 := merror.WrapCodef(errors.New("original"), mcode.CodeServerBusy, "server %s busy", "auth")
	assert.NotNil(t, err4)
	assert.Equal(t, "server auth busy: original", err4.Error())
	assert.Equal(t, mcode.CodeServerBusy, merror.Code(err4))

	// Code on wrapped error
	err5 := merror.Wrap(err4, "outer layer")
	assert.Equal(t, mcode.CodeServerBusy, merror.Code(err5))

	// Nil cases
	assert.Nil(t, merror.WrapCode(nil, mcode.CodeOK))
	assert.Nil(t, merror.WrapCodef(nil, mcode.CodeOK, "format"))
	assert.Equal(t, mcode.CodeNil, merror.Code(nil))
}

// Test for Stack, Cause, Current, Unwrap
func TestStack(t *testing.T) {
	// Cause
	err1 := errors.New("root")
	err2 := merror.Wrap(err1, "wrapped")
	assert.Equal(t, err1, merror.Cause(err2))
	assert.Equal(t, err1, merror.Cause(err1))

	// Stack
	stackStr := merror.Stack(err2)
	assert.True(t, strings.Contains(stackStr, "merror_test.go"))
	assert.True(t, strings.Contains(stackStr, "TestStack"))

	// Current
	currentErr := merror.Current(err2)
	assert.NotNil(t, currentErr)
	assert.Equal(t, "wrapped", currentErr.Error())

	// Unwrap
	unwrappedErr := merror.Unwrap(err2)
	assert.Equal(t, err1, unwrappedErr)
	assert.Nil(t, merror.Unwrap(err1))

	// Nil cases
	assert.Nil(t, merror.Cause(nil))
	assert.Equal(t, "", merror.Stack(nil))
	assert.Nil(t, merror.Current(nil))
	assert.Nil(t, merror.Unwrap(nil))

	// HasStack
	assert.True(t, merror.HasStack(merror.New("has stack")))
	assert.False(t, merror.HasStack(errors.New("no stack")))
}

// Test for Equal, Is, As
func TestComparison(t *testing.T) {
	err1 := merror.NewCode(mcode.CodeNotFound, "not found")
	err2 := merror.NewCode(mcode.CodeNotFound, "not found")
	err3 := merror.NewCode(mcode.CodeForbidden, "forbidden")

	// Equal
	assert.True(t, merror.Equal(err1, err2))
	assert.False(t, merror.Equal(err1, err3))
	assert.True(t, merror.Equal(nil, nil))
	assert.False(t, merror.Equal(err1, nil))

	// Equal with different types
	err4 := errors.New("not found")
	assert.False(t, merror.Equal(err1, err4))
	assert.False(t, merror.Equal(err4, err1))

	// Is
	wrapped := merror.Wrap(err1, "wrapped")
	assert.True(t, errors.Is(wrapped, err1))
	assert.False(t, errors.Is(wrapped, err3))

	// As
	var e *merror.Error
	assert.True(t, errors.As(wrapped, &e))
	assert.NotNil(t, e)
	assert.Equal(t, "wrapped: not found", e.Error())
}

// Test for methods on Error struct
func TestErrorMethods(t *testing.T) {
	root := errors.New("root error")
	mid := merror.WrapCode(root, mcode.CodeInternalError, "mid layer")
	outer := merror.WrapCode(mid, mcode.CodeServerBusy, "outer layer")

	// Error()
	assert.Equal(t, "outer layer: mid layer: root error", outer.Error())

	// Cause()
	merr, ok := outer.(merror.ICause)
	assert.True(t, ok)
	assert.Equal(t, root, merr.Cause())

	// Unwrap()
	unwrapped := merror.Unwrap(outer)
	assert.Equal(t, mid, unwrapped)

	// Code()
	codeProvider, ok := outer.(merror.ICode)
	assert.True(t, ok)
	assert.Equal(t, mcode.CodeServerBusy, codeProvider.Code())
	// Test code propagation from inner error
	midCodeProvider, ok := merror.Unwrap(outer).(merror.ICode)
	assert.True(t, ok)
	assert.Equal(t, mcode.CodeInternalError, midCodeProvider.Code())

	// Format()
	// %s
	assert.Equal(t, "outer layer: mid layer: root error", fmt.Sprintf("%s", outer))
	// %+v
	stackPlusV := fmt.Sprintf("%+v", outer)
	assert.Contains(t, stackPlusV, "outer layer: mid layer: root error")
	assert.Contains(t, stackPlusV, "merror_test.go")
	// %v
	assert.Equal(t, "outer layer: mid layer: root error", fmt.Sprintf("%v", outer))
	// %-v
	assert.Equal(t, "outer layer: mid layer: root error", fmt.Sprintf("%-v", outer))

	// Test *Error methods with nil receiver
	var nilErr *merror.Error
	assert.Equal(t, "", nilErr.Error())
	assert.Nil(t, nilErr.Cause())
	assert.Nil(t, nilErr.Unwrap())
	assert.Nil(t, nilErr.Current())
	assert.Equal(t, mcode.CodeNil, nilErr.Code())
	nilErr.SetCode(mcode.CodeOK) // Should not panic
	assert.Equal(t, "", nilErr.Stack())

	// Test (*Error).Error() with empty text but valid code
	errWithCodeOnly := merror.NewCode(mcode.CodeOK)
	assert.Equal(t, "OK", errWithCodeOnly.Error())

	// Test (*Error).Cause() with a standard error wrapped inside
	wrappedStdErr := merror.Wrap(errors.New("std error"), "wrapped")
	assert.Equal(t, errors.New("std error"), merror.Cause(wrappedStdErr))

	// Test (*Error).SetCode
	errToSetCode := merror.New("err")
	asMerror, ok := errToSetCode.(*merror.Error)
	assert.True(t, ok)
	asMerror.SetCode(mcode.CodeServerBusy)
	assert.Equal(t, mcode.CodeServerBusy, merror.Code(errToSetCode))
}

// Test for JSON Marshaling
func TestJson(t *testing.T) {
	err := merror.New("json test")
	b, e := json.Marshal(err)
	assert.Nil(t, e)
	assert.Equal(t, `"json test"`, string(b))
}
