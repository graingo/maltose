package controller

import (
	"context"
)

type HelloController struct{}

func (h *HelloController) Hello(ctx context.Context, req *v1.HelloReq) (*v1.HelloRes, error) {
	return &HelloRes{Name: req.Name}, nil
}
