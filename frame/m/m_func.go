package m

import (
	"context"

	"github.com/mingzaily/maltose/net/mhttp"
)

func RequestFromCtx(ctx context.Context) *mhttp.Request {
	return mhttp.RequestFromCtx(ctx)
}
