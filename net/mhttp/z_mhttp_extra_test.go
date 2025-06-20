package mhttp_test

import (
	"net/http"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPProf(t *testing.T) {
	teardown := setupServer(t, func(s *mhttp.Server) {
		s.EnablePProf()
	})
	defer teardown()

	pprofPaths := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	}

	// Special case for profile, as it blocks for 30s by default
	t.Run("/debug/pprof/profile", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/debug/pprof/profile?seconds=1")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	for _, path := range pprofPaths {
		t.Run(path, func(t *testing.T) {
			resp, err := http.Get(baseURL + path)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
