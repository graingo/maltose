package mhttp

import (
	"context"
	"net/http"
)

func (s *Server) registerHealthCheck(ctx context.Context) {
	if s.config.HealthCheck == "" {
		return
	}

	s.GET(s.config.HealthCheck, func(r *Request) {
		r.JSON(http.StatusOK, map[string]any{
			"status": "ok",
		})
	})

	s.logger().Infof(ctx, "Health check endpoint registered at %s", s.config.HealthCheck)
}
