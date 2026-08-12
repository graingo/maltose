package merror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedError struct{ message string }

func (e *typedError) Error() string { return e.message }

func TestStandardErrorChainHelpers(t *testing.T) {
	target := &typedError{message: "target"}
	wrapped := errors.Join(errors.New("other"), target)
	assert.True(t, Is(wrapped, target))
	assert.False(t, Is(wrapped, errors.New("missing")))

	var matched *typedError
	require.True(t, As(wrapped, &matched))
	assert.Same(t, target, matched)
}
