package gins

import (
	"github.com/mingzaily/maltose/os/gcfg"
)

// Config 返回一个 gcfg.Config 实例
func Config(name ...string) *gcfg.Config {
	return gcfg.Instance(name...)
}
