package mcode_test

import (
	"fmt"
	"testing"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/stretchr/testify/assert"
)

func TestPredefinedCodes(t *testing.T) {
	assert.Equal(t, 0, mcode.CodeOK.Code())
	assert.Equal(t, "OK", mcode.CodeOK.Message())
	assert.Nil(t, mcode.CodeOK.Detail())

	assert.Equal(t, 1003, mcode.CodeValidationFailed.Code())
	assert.Equal(t, "Validation Failed", mcode.CodeValidationFailed.Message())

	assert.Equal(t, -1, mcode.CodeNil.Code())
}

func TestNew(t *testing.T) {
	c := mcode.New(99, "Custom Error", "some detail")
	assert.Equal(t, 99, c.Code())
	assert.Equal(t, "Custom Error", c.Message())
	assert.Equal(t, "some detail", c.Detail())
}

func TestWithCode(t *testing.T) {
	// Create a new code with different detail from a predefined code
	c := mcode.WithCode(mcode.CodeInvalidParameter, "field: username")
	assert.Equal(t, mcode.CodeInvalidParameter.Code(), c.Code())
	assert.Equal(t, mcode.CodeInvalidParameter.Message(), c.Message())
	assert.Equal(t, "field: username", c.Detail())
}

func TestLocalCode_String(t *testing.T) {
	// Test with detail
	c1 := mcode.New(1001, "Invalid Parameter", "username")
	assert.Equal(t, "1001:Invalid Parameter username", fmt.Sprintf("%s", c1))

	// Test without detail
	c2 := mcode.New(1004, "Not Found", nil)
	assert.Equal(t, "1004:Not Found", fmt.Sprintf("%s", c2))

	// Test without message and detail
	c3 := mcode.New(500, "", nil)
	assert.Equal(t, "500", fmt.Sprintf("%s", c3))
}
