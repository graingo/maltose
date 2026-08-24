# Nacos configuration adapter

Package `nacos` connects Maltose `mcfg.Config` to Nacos Config and optionally keeps the in-memory value synchronized with remote changes.

## Installation

```bash
go get github.com/graingo/maltose/contrib/config/nacos
```

## Usage

```go
package main

import (
	"context"

	"github.com/graingo/maltose/contrib/config/nacos"
	"github.com/graingo/maltose/os/mcfg"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func loadConfig(ctx context.Context) (*mcfg.Config, error) {
	adapter, err := nacos.New(ctx, nacos.Config{
		ServerConfigs: []constant.ServerConfig{
			{IpAddr: "127.0.0.1", Port: 8848},
		},
		ClientConfig: constant.ClientConfig{
			CacheDir: "/tmp/nacos/cache",
			LogDir:   "/tmp/nacos/log",
		},
		ConfigParam: vo.ConfigParam{
			DataId: "app.yaml",
			Group:  "DEFAULT_GROUP",
			Type:   "yaml",
		},
		Watch: true,
	})
	if err != nil {
		return nil, err
	}

	return mcfg.NewWithAdapter(adapter), nil
}
```

The remote document must have an object at its root. `Watch` updates the adapter's in-memory value when Nacos publishes a change. If process-wide `mcfg` hooks are registered, call `Config.ClearCache` from `OnConfigChange` so the hooks run again for the new value.

To replace the process-wide default configuration, initialize the adapter before other components and call `m.Config().SetAdapter(adapter)`. Prefer `m.NewScope(mcfg.NewWithAdapter(adapter))` when an application or test needs an isolated component scope.

## License

This module is licensed under the [MIT License](../../../LICENSE).
