# apollo

Package `apollo` implements maltose `mcfg.Adapter` using apollo service.

## Installation

```bash
go get -u github.com/graingo/maltose/contrib/config/apollo
```

## Usage

### Create a custom boot package

If you wish using configuration from apollo globally,
it is strongly recommended creating a custom boot package in very top import,
which sets the Adapter of default configuration instance before any other package boots.

```go
package boot

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/contrib/config/apollo"
	"github.com/graingo/maltose/frame/m"
)

func init() {
	var (
		ctx     = context.Background()
		appId   = "SampleApp"
		cluster = "default"
		ip      = "http://localhost:8080"
	)
	// Create apollo Client that implements mcfg.Adapter
	adapter, err := apollo.New(ctx, apollo.Config{
		AppID:   appId,
		IP:      ip,
		Cluster: cluster,
	})
	if err != nil {
		fmt.Printf("init apollo failed: %+v\n", err)
		return
	}
	// Change the adapter of default configuration instance
	m.Config().SetAdapter(adapter)
}
```

## License

`maltose apollo` is licensed under the [MIT License](../../../LICENSE), 100% free and open-source, forever.
