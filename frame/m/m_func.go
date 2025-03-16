package m

import (
	"context"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/net/mhttp"
)

func RequestFromCtx(ctx context.Context) *mhttp.Request {
	return mhttp.RequestFromCtx(ctx)
}

func NewVar(value interface{}, safe ...bool) *Var {
	return mvar.New(value, safe...)
}
