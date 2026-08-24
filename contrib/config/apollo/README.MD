# Apollo configuration adapter

Package `apollo` connects Maltose `mcfg.Config` to Apollo Config and optionally keeps the in-memory value synchronized with remote changes.

## Installation

```bash
go get github.com/graingo/maltose/contrib/config/apollo
```

## Usage

```go
package main

import (
	"context"

	"github.com/graingo/maltose/contrib/config/apollo"
	"github.com/graingo/maltose/os/mcfg"
)

func loadConfig(ctx context.Context) (*mcfg.Config, error) {
	adapter, err := apollo.New(ctx, apollo.Config{
		AppID:         "checkout-api",
		IP:            "http://127.0.0.1:8080",
		Cluster:       "default",
		NamespaceName: "application",
		Watch:         true,
	})
	if err != nil {
		return nil, err
	}

	return mcfg.NewWithAdapter(adapter), nil
}
```

`Watch` refreshes the adapter's in-memory value when Apollo publishes a change. If process-wide `mcfg` hooks are registered, clear the `Config` cache after a remote change so the hooks run again for the new value.

To replace the process-wide default configuration, initialize the adapter before other components and call `m.Config().SetAdapter(adapter)`. Prefer `m.NewScope(mcfg.NewWithAdapter(adapter))` when an application or test needs an isolated component scope.

## License

This module is licensed under the [MIT License](../../../LICENSE).
