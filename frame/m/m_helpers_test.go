package m

import (
	"context"
	"testing"

	"github.com/graingo/maltose/frame/mins"
	"github.com/graingo/maltose/os/mcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameworkHelperDelegation(t *testing.T) {
	assert.Nil(t, RequestFromCtx(nil))
	assert.Nil(t, RequestFromCtx(context.Background()))
	assert.Equal(t, 42, NewVar(42).Int())
	assert.Equal(t, "value", NewVar("value", true).String())

	adapter, err := mcfg.NewAdapterContent("logger:\n  service_name: helper-test\n", "yaml")
	require.NoError(t, err)
	scope := NewScope(mcfg.NewWithAdapter(adapter))
	assert.Equal(t, "helper-test", scope.Log().GetConfig().ServiceName)
	assert.Same(t, mins.DefaultScope(), DefaultScope())

	assert.Same(t, mins.Config(), Config())
	assert.Same(t, mins.Log("m-helper"), Log("m-helper"))
	assert.Same(t, mins.Server("m-helper"), Server("m-helper"))
}
