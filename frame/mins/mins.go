package mins

import (
	"github.com/graingo/maltose/container/minstance"
	"github.com/graingo/maltose/os/mcfg"
	"github.com/graingo/maltose/os/mlog"
)

const (
	frameCoreNameLogger = "maltose.logger"
	frameCoreNameRedis  = "maltose.redis"
	frameCoreNameServer = "maltose.server"
	frameCoreNameDB     = "maltose.db"
)

// Scope owns framework instances for one application or test boundary.
// Package-level helpers use the default scope for backward compatibility.
type Scope struct {
	config              *mcfg.Config
	dbInstances         *minstance.Container
	redisInstances      *minstance.Container
	serverInstances     *minstance.Container
	loggerInstances     *minstance.Container
	useGlobalComponents bool
}

var defaultScope = newScope(nil, true)

// NewScope creates an isolated instance scope backed by config.
// The config must be non-nil so a scope never falls back to process-global configuration implicitly.
func NewScope(config *mcfg.Config) *Scope {
	if config == nil {
		panic("mins: scope config must not be nil")
	}
	return newScope(config, false)
}

// DefaultScope returns the process-wide scope used by package-level helpers.
func DefaultScope() *Scope {
	return defaultScope
}

func newScope(config *mcfg.Config, useGlobalComponents bool) *Scope {
	return &Scope{
		config:              config,
		dbInstances:         minstance.New(),
		redisInstances:      minstance.New(),
		serverInstances:     minstance.New(),
		loggerInstances:     minstance.New(),
		useGlobalComponents: useGlobalComponents,
	}
}

func (s *Scope) configInstance() *mcfg.Config {
	if s == nil {
		panic("mins: scope must not be nil")
	}
	if s.config != nil {
		return s.config
	}
	return mcfg.Instance()
}

func (s *Scope) newLogger(instanceName string) *mlog.Logger {
	if s.useGlobalComponents {
		return mlog.Instance(instanceName)
	}
	return mlog.New()
}
