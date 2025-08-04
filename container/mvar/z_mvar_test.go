package mvar_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/graingo/maltose/container/mvar"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test with a value and default safety (false)
	v := mvar.New("hello")
	assert.NotNil(t, v)
	assert.Equal(t, "hello", v.Val())

	// Test with a value and safe = false
	v = mvar.New("hello", false)
	assert.NotNil(t, v)
	assert.Equal(t, "hello", v.Val())

	// Test with a value and safe = true
	v = mvar.New("hello", true)
	assert.NotNil(t, v)
	assert.Equal(t, "hello", v.Val())
}

func TestVar_Val(t *testing.T) {
	v := mvar.New("test")
	assert.Equal(t, "test", v.Val())
	assert.Equal(t, "test", v.Interface())

	var nilVar *mvar.Var
	assert.Nil(t, nilVar.Val())
	assert.Nil(t, nilVar.Interface())
}

func TestVar_Conversions(t *testing.T) {
	// String
	assert.Equal(t, "123", mvar.New(123).String())
	assert.Equal(t, "true", mvar.New(true).String())
	assert.Equal(t, "", (*mvar.Var)(nil).String())

	// Bool
	assert.True(t, mvar.New("true").Bool())
	assert.False(t, mvar.New("false").Bool())
	assert.False(t, mvar.New(0).Bool())
	assert.True(t, mvar.New(1).Bool())
	assert.False(t, (*mvar.Var)(nil).Bool())

	// Int
	assert.Equal(t, 123, mvar.New("123").Int())
	assert.Equal(t, 0, (*mvar.Var)(nil).Int())

	// Int64
	assert.Equal(t, int64(123), mvar.New("123").Int64())
	assert.Equal(t, int64(0), (*mvar.Var)(nil).Int64())

	// Uint64
	assert.Equal(t, uint64(123), mvar.New("123").Uint64())
	assert.Equal(t, uint64(0), (*mvar.Var)(nil).Uint64())

	// Float64
	assert.Equal(t, 123.45, mvar.New("123.45").Float64())
	assert.Equal(t, 0.0, (*mvar.Var)(nil).Float64())
}

func TestVar_Time(t *testing.T) {
	ts := "2024-01-02T15:04:05Z"
	tVal, _ := time.Parse(time.RFC3339, ts)
	v := mvar.New(ts)
	assert.Equal(t, tVal, v.Time(time.RFC3339))
	assert.True(t, (*mvar.Var)(nil).Time().IsZero())

	now := time.Now()
	v = mvar.New(now)
	assert.Equal(t, now.Unix(), v.Time().Unix())
}

func TestVar_Struct(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	data := map[string]any{
		"Name": "John",
		"Age":  30,
	}
	v := mvar.New(data)
	var user User
	err := v.Struct(&user)
	assert.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 30, user.Age)

	err = (*mvar.Var)(nil).Struct(&user)
	assert.NoError(t, err)
}

func TestVar_JSON(t *testing.T) {
	// Marshal
	v := mvar.New(map[string]any{"key": "value"})
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"key":"value"}`, string(b))

	// Unmarshal
	var v2 mvar.Var
	err = json.Unmarshal([]byte(`{"key":"value"}`), &v2)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"key": "value"}, v2.Val())
}

func TestVar_IsType(t *testing.T) {
	assert.True(t, (*mvar.Var)(nil).IsNil())
	assert.True(t, mvar.New(nil).IsNil())
	assert.False(t, mvar.New("").IsNil())

	assert.True(t, mvar.New("").IsEmpty())
	assert.True(t, mvar.New(0).IsEmpty())
	assert.True(t, mvar.New(false).IsEmpty())
	assert.False(t, mvar.New("a").IsEmpty())

	assert.True(t, mvar.New(1).IsInt())
	assert.False(t, mvar.New(1.1).IsInt())

	assert.True(t, mvar.New(uint(1)).IsUint())
	assert.False(t, mvar.New(-1).IsUint())

	assert.True(t, mvar.New(1.1).IsFloat())
	assert.False(t, mvar.New(1).IsFloat())

	assert.True(t, mvar.New([]int{1}).IsSlice())
	assert.False(t, mvar.New("a").IsSlice())

	assert.True(t, mvar.New(map[string]any{}).IsMap())
	assert.False(t, mvar.New("a").IsMap())

	type MyStruct struct{}
	assert.True(t, mvar.New(MyStruct{}).IsStruct())
	assert.False(t, mvar.New("a").IsStruct())
}

func TestVar_Map(t *testing.T) {
	// From map[any]any
	m1 := map[any]any{1: "one", "two": 2}
	v1 := mvar.New(m1)
	assert.Equal(t, map[string]any{"1": "one", "two": 2}, v1.Map())

	// From struct
	type User struct {
		Name string
		Age  int
	}
	u := User{Name: "John", Age: 30}
	v2 := mvar.New(u)
	assert.Equal(t, map[string]any{"Name": "John", "Age": 30}, v2.Map())

	assert.Nil(t, (*mvar.Var)(nil).Map())
}

func TestVar_Set(t *testing.T) {
	// Unsafe
	v := mvar.New("initial")
	old := v.Set("new")
	assert.Equal(t, "initial", old)
	assert.Equal(t, "new", v.Val())

	// Safe
	v = mvar.New("initial", true)
	old = v.Set("new")
	assert.Equal(t, "initial", old)
	assert.Equal(t, "new", v.Val())
}
