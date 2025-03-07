package mins

import (
	"github.com/graingo/maltose/os/mcfg"
)

// Config 返回一个 gcfg.Config 实例
func Config(name ...string) *mcfg.Config {
	return mcfg.Instance(name...)
}
