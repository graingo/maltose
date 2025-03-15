package mins

import (
	"github.com/graingo/maltose/os/mcfg"
)

// Config returns a gcfg.Config instance
func Config(name ...string) *mcfg.Config {
	return mcfg.Instance(name...)
}
