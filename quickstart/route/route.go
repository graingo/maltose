package route

import (
	"quickstart/controller"

	"github.com/mingzaily/maltose/net/mhttp"
)

func Build(s *mhttp.Server) {
	// s.Use(mhttp.MiddlewareResponse())
	s.Use(mhttp.MiddlewareLog())

	hello := controller.NewHelloController()

	g := s.Group("api/v1")
	g.Bind(hello)
}
