package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRejectsUnsupportedFormat(t *testing.T) {
	err := Generate(".", "openapi.toml", "toml")
	assert.EqualError(t, err, `unsupported OpenAPI output format "toml": use yaml or json`)
}
