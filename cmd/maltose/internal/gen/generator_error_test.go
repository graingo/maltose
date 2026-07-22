package gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogicGeneratorPropagatesParseErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "broken.go")
	require.NoError(t, os.WriteFile(file, []byte("package broken\nfunc {"), 0o644))
	generator := &LogicGenerator{ModuleName: "example.com/project"}
	_, err := generator.genFromFile(file)
	assert.Error(t, err)
}

func TestGenerateFileDoesNotWriteInvalidGo(t *testing.T) {
	output := filepath.Join(t.TempDir(), "generated.go")
	err := generateFile(output, "invalid", "package generated\nfunc {", nil)
	assert.Error(t, err)
	_, statErr := os.Stat(output)
	assert.True(t, os.IsNotExist(statErr))
}
