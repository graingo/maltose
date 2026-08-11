package otlpmetric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func TestCreateResourceIncludesConfiguredAttributes(t *testing.T) {
	options := defaultOptions()
	WithServiceName("checkout-api")(&options)
	WithServiceVersion("v1.2.0")(&options)
	WithEnvironment("staging")(&options)
	WithResourceAttribute("tenant", "malt")(&options)

	resource, err := createResource(options)
	require.NoError(t, err)

	attributes := resource.Set()
	serviceName, ok := attributes.Value(semconv.ServiceNameKey)
	require.True(t, ok)
	assert.Equal(t, "checkout-api", serviceName.AsString())

	serviceVersion, ok := attributes.Value(semconv.ServiceVersionKey)
	require.True(t, ok)
	assert.Equal(t, "v1.2.0", serviceVersion.AsString())

	environment, ok := attributes.Value(semconv.DeploymentEnvironmentKey)
	require.True(t, ok)
	assert.Equal(t, "staging", environment.AsString())

	tenant, ok := attributes.Value("tenant")
	require.True(t, ok)
	assert.Equal(t, "malt", tenant.AsString())
}
