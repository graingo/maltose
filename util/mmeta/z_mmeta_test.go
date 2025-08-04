package mmeta_test

import (
	"encoding/json"
	"testing"

	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeta(t *testing.T) {
	type A struct {
		mmeta.Meta `tag:"123" orm:"456"`
		ID         int
		Name       string
	}

	a := &A{
		ID:   100,
		Name: "john",
	}

	t.Run("basic_operations", func(t *testing.T) {
		// Test mmeta.Data()
		data := mmeta.Data(a)
		assert.Len(t, data, 2, "should contain 2 meta items")
		assert.Equal(t, "123", data["tag"], "tag value should be correct")
		assert.Equal(t, "456", data["orm"], "orm value should be correct")

		// Test mmeta.Get()
		assert.Equal(t, "123", mmeta.Get(a, "tag").String(), "Get should retrieve correct tag value")
		assert.Equal(t, "456", mmeta.Get(a, "orm").String(), "Get should retrieve correct orm value")
	})

	t.Run("json_marshaling", func(t *testing.T) {
		b, err := json.Marshal(a)
		require.NoError(t, err, "JSON marshaling should not produce an error")

		// The mmeta.Meta field should be ignored during marshaling.
		expectedJSON := `{"ID":100,"Name":"john"}`
		assert.JSONEq(t, expectedJSON, string(b), "JSON output should not include meta tags")
	})

	t.Run("edge_cases", func(t *testing.T) {
		// Test getting a non-existent key
		assert.Nil(t, mmeta.Get(a, "non_existent_key"), "getting a non-existent key should return nil")

		// Test with a nil pointer
		var nilA *A
		assert.Equal(t, map[string]string{}, mmeta.Data(nilA), "Data() on a nil pointer should return empty map")
		assert.Nil(t, mmeta.Get(nilA, "tag"), "Get() on a nil pointer should return nil")

		// Test with a struct without the Meta field
		type B struct {
			ID int
		}
		b := &B{ID: 200}
		assert.Equal(t, map[string]string{}, mmeta.Data(b), "Data() on a struct without Meta should return empty map")
		assert.Nil(t, mmeta.Get(b, "tag"), "Get() on a struct without Meta should return nil")
	})
}
