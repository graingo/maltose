package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorCommandsRejectPositionalArguments(t *testing.T) {
	commands := []struct {
		name string
		run  func() error
	}{
		{name: "model", run: func() error { return modelCmd.Args(modelCmd, []string{"unexpected"}) }},
		{name: "dao", run: func() error { return daoCmd.Args(daoCmd, []string{"unexpected"}) }},
		{name: "service", run: func() error { return serviceCmd.Args(serviceCmd, []string{"unexpected"}) }},
		{name: "logic", run: func() error { return logicCmd.Args(logicCmd, []string{"unexpected"}) }},
		{name: "openapi", run: func() error { return openapiCmd.Args(openapiCmd, []string{"unexpected"}) }},
	}

	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) {
			assert.Error(t, command.run())
		})
	}
}

func TestValidateServiceMode(t *testing.T) {
	assert.NoError(t, validateServiceMode("", "interface"))
	assert.NoError(t, validateServiceMode("", "struct"))
	assert.EqualError(t, validateServiceMode("", "typo"), `unsupported service generation mode "typo": use interface or struct`)
	assert.NoError(t, validateServiceMode("user", "typo"), "mode is ignored when a service name is provided")
}
