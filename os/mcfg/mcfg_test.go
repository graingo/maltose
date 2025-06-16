package mcfg_test

import (
	"context"
	"testing"

	"github.com/graingo/maltose/os/mcfg"
)

var (
	ctx           = context.Background()
	fixtureConfig = `
app:
  name: "test-app"
  version: "1.0.0"
  debug: true
  port: 8080
database:
  host: "localhost"
  port: 5432
  settings:
    timeout: 30
    maxConn: 100
`
	fixtureConfigTest = `
app:
  name: "test-app-test"
  version: "2.0.0"
`
)

func TestConfig_Basic(t *testing.T) {
	c, err := mcfg.New()
	if err != nil {
		t.Fatal("New() returned error")
	}
	if c == nil {
		t.Fatal("New() returned nil")
	}

	// 测试空配置
	val, err := c.Get(ctx, "any.key")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if val != nil {
		t.Errorf("Get() expected nil, got %v", val)
	}

	val, err = c.Get(ctx, "test.key")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if val.String() != "value" {
		t.Errorf("Get() expected 'value', got %v", val)
	}
}

func TestConfig_Adapter(t *testing.T) {
	c, err := mcfg.New()
	if err != nil {
		t.Fatal("New() returned error")
	}
	if c == nil {
		t.Fatal("New() returned nil")
	}

	// 测试设置适配器
	mockAdapter := newMockAdapter()
	c.SetAdapter(mockAdapter)

	// 测试通过适配器获取数据
	val, err := c.Get(ctx, "mock.key")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if val.String() != "mock_value" {
		t.Errorf("Get() expected 'mock_value', got %v", val)
	}
}

// 辅助测试的 mock 适配器
type mockAdapter struct{}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{}
}

func (m *mockAdapter) Get(_ context.Context, pattern string) (interface{}, error) {
	return "mock_value", nil
}

func (m *mockAdapter) Set(_ context.Context, pattern string, value interface{}) error {
	return nil
}

func (m *mockAdapter) Data(_ context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"mock.key": "mock_value",
	}, nil
}

func (m *mockAdapter) Available(_ context.Context, _ ...string) bool {
	return true
}

func (m *mockAdapter) MergeConfigMap(_ context.Context, _ map[string]any) error {
	return nil
}
