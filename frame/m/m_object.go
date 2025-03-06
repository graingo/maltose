package m

import (
	"github.com/savorelle/maltose/frame/mins"
	"github.com/savorelle/maltose/net/mhttp"
	"github.com/savorelle/maltose/os/mcfg"
	"github.com/savorelle/maltose/os/mlog"
)

// Server 返回指定名称的 HTTP 服务器实例
func Server(name ...interface{}) *mhttp.Server {
	return mins.Server(name...)
}

// Config 返回具有指定名称的配置对象的实例
func Config(name ...string) *mcfg.Config {
	return mins.Config(name...)
}

// Log 返回具有指定名称的日志对象的实例
func Log(name ...string) *mlog.Logger {
	return mins.Log(name...)
}
