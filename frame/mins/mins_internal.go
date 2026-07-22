package mins

import (
	"fmt"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
)

func mustConfigMap(value any, node string) map[string]any {
	configMap, ok := value.(map[string]any)
	if !ok {
		panic(merror.NewCodef(
			mcode.CodeInvalidConfiguration,
			"configuration node %q must be an object, got %s",
			node,
			fmt.Sprintf("%T", value),
		))
	}
	return configMap
}
