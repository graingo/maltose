package mhttp

import (
	"context"
	"net/http"
)

func (s *Server) registerHealthCheck(ctx context.Context) {
	if !s.config.HealthCheck {
		return
	}

	s.GET("/health", func(r *Request) {
		r.JSON(http.StatusOK, map[string]any{
			"status": "ok",
		})
	})
	s.logger().Infof(ctx, "Health check endpoint registered at /health")
}
