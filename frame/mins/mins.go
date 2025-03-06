package mins

import "github.com/savorelle/maltose/container/minstance"

const (
	frameCoreNameLogger = "maltose.logger"
	frameCoreNameRedis  = "maltose.redis"
	frameCoreNameServer = "maltose.server"
)

var (
	globalInstances = minstance.New()
)
