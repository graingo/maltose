# nacos

Package `nacos` implements GoFrame `gcfg.Adapter` using nacos service.

# Installation

```
go get -u github.com/gogf/gf/contrib/config/nacos/v2
```

# Usage

## Create a custom boot package

If you wish using configuration from nacos globally,
it is strongly recommended creating a custom boot package in very top import,
which sets the Adapter of default configuration instance before any other package boots.

```go
package boot

import (
	"github.com/gogf/gf/contrib/config/nacos/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func init() {
	var (
		ctx          = gctx.GetInitCtx()
		serverConfig = constant.ServerConfig{
			IpAddr: "localhost",
			Port:   8848,
		}
		clientConfig = constant.ClientConfig{
			CacheDir: "/tmp/nacos",
			LogDir:   "/tmp/nacos",
		}
		configParam = vo.ConfigParam{
			DataId: "config.toml",
			Group:  "test",
		}
	)
	// Create anacosClient that implements gcfg.Adapter.
	adapter, err := nacos.New(ctx, nacos.Config{
		ServerConfigs: []constant.ServerConfig{serverConfig},
		ClientConfig:  clientConfig,
		ConfigParam:   configParam,
	})
	if err != nil {
		g.Log().Fatalf(ctx, `%+v`, err)
	}
	// Change the adapter of default configuration instance.
	g.Cfg().SetAdapter(adapter)
}
```

## License

`GoFrame nacos` is licensed under the [MIT License](../../../LICENSE), 100% free and open-source, forever.
