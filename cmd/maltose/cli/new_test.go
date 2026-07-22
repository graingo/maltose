package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectTargetStaysInsideWorkingDirectory(t *testing.T) {
	cwd := t.TempDir()
	target, err := projectTarget(cwd, "services/example")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(cwd, "services", "example"), target)

	for _, invalid := range []string{"", ".", "..", "../outside", filepath.Join(string(filepath.Separator), "tmp", "outside")} {
		_, err := projectTarget(cwd, invalid)
		assert.Error(t, err, invalid)
	}
}
