package mins

import "github.com/graingo/maltose/container/minstance"

const (
	frameCoreNameLogger = "maltose.logger"
	frameCoreNameRedis  = "maltose.redis"
	frameCoreNameServer = "maltose.server"
	frameCoreNameDB     = "maltose.db"
)

var (
	globalInstances = minstance.New()
)
