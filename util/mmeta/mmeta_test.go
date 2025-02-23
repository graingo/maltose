package mmeta_test

import (
	"encoding/json"
	"testing"

	"github.com/mingzaily/maltose/util/mmeta"
)

func TestMeta_Basic(t *testing.T) {
	type A struct {
		mmeta.Meta `tag:"123" orm:"456"`
		Id         int
		Name       string
	}

	a := &A{
		Id:   100,
		Name: "john",
	}

	// 测试基础功能
	if len(mmeta.Data(a)) != 2 {
		t.Error("Expected 2 meta items")
	}
	if mmeta.Get(a, "tag").String() != "123" {
		t.Error("Expected tag value '123'")
	}
	if mmeta.Get(a, "orm").String() != "456" {
		t.Error("Expected orm value '456'")
	}
	if mmeta.Get(a, "none") != nil {
		t.Error("Expected nil for non-existent key")
	}

	// 测试 JSON 序列化
	b, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}
	if string(b) != `{"Id":100,"Name":"john"}` {
		t.Error("Unexpected JSON result")
	}
}

func TestMeta_Convert_Map(t *testing.T) {
	type A struct {
		mmeta.Meta `tag:"123" orm:"456"`
		Id         int
		Name       string
	}

	a := &A{
		Id:   100,
		Name: "john",
	}
	m := mmeta.Data(a)
	if len(m) != 2 {
		t.Error("Expected 2 meta items")
	}
	if m["Meta"] != "" {
		t.Error("Unexpected Meta field in map")
	}
}

func TestMeta_Json(t *testing.T) {
	type A struct {
		mmeta.Meta `tag:"123" orm:"456"`
		Id         int
	}

	a := &A{
		Id: 100,
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}
	if string(b) != `{"Id":100}` {
		t.Error("Unexpected JSON result")
	}
}
