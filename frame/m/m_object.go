package m

import (
	"context"

	"github.com/graingo/maltose/database/mdb"
	"github.com/graingo/maltose/frame/mins"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/os/mcfg"
	"github.com/graingo/maltose/os/mlog"
)

// Server returns the instance of the HTTP server with the specified name.
func Server(name ...string) *mhttp.Server {
	return mins.Server(name...)
}

// Config returns the instance of the configuration with the specified name.
func Config(name ...string) *mcfg.Config {
	return mins.Config(name...)
}

// Log returns the instance of the logger with the specified name.
func Log(name ...string) *mlog.Logger {
	return mins.Log(name...)
}

// DB returns the instance of the database with the specified name.
func DB(name ...string) *mdb.DB {
	return mins.DB(name...)
}

// DBContext returns the instance of the database with the specified name and context.
func DBContext(ctx context.Context, name ...string) *mdb.DB {
	return mins.DB(name...).WithContext(ctx)
}
