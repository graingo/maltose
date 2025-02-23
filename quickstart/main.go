package main

import (
	"context"

	"github.com/mingzaily/maltose/frame/m"
	"github.com/mingzaily/maltose/os/mcfg"
)

type HelloReq struct {
	m.Meta `method:"GET" path:"/api/v1/hello" summary:"Hello请求"`
	Name   string `form:"name"`
}

type HelloRes struct {
	Name string `json:"name"`
}

type HelloController struct{}

func (h *HelloController) Hello(ctx context.Context, req *HelloReq) (*HelloRes, error) {
	return &HelloRes{Name: req.Name}, nil
}

func main() {
	adapter, err := mcfg.NewAdapterFile()
	if err != nil {
		panic(err)
	}
	adapter.SetFileName("dev")
	m.Config().SetAdapter(adapter)

	s := m.Server()
	s.BindHandler("GET", "/hello", func(c *mhttp.Request) {
		c.JSON(200, "Hello, World!")
	})
	s.Run()
}
