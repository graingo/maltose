package mins

import (
	"github.com/graingo/maltose/os/mcfg"
)

// Config returns an mcfg.Config instance.
func Config(name ...string) *mcfg.Config {
	return mcfg.Instance(name...)
}

// Config returns the configuration source owned by the scope.
func (s *Scope) Config() *mcfg.Config {
	return s.configInstance()
}
