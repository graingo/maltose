package controller

import (
	"context"
	v1 "quickstart/api/v1"
)

type HelloController struct{}

func NewHelloController() *HelloController {
	return &HelloController{}
}

func (h *HelloController) Hello(ctx context.Context, req *v1.HelloReq) (*v1.HelloRes, error) {
	return &v1.HelloRes{Name: req.Name}, nil
}
