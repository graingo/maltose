package mhttp_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtraFeatures(t *testing.T) {
	t.Run("pprof_endpoints", func(t *testing.T) {
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

		// Special case for profile, as it blocks for 30s by default.
		// We run it with a very short duration.
		t.Run("profile_endpoint", func(t *testing.T) {
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
	})

	t.Run("custom_health_check", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			// Explicitly enable health check for this test
			s.SetConfigWithMap(map[string]any{
				"health_check": "/healthz", // Use a custom path
			})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/healthz")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"status":"ok"`)
	})

	t.Run("static_file_serving", func(t *testing.T) {
		// Create a dummy file to serve
		tmpDir := t.TempDir()
		filePath := tmpDir + "/test.txt"
		fileContent := "hello world from static file"
		err := ioutil.WriteFile(filePath, []byte(fileContent), 0644)
		require.NoError(t, err)

		teardown := setupServer(t, func(s *mhttp.Server) {
			s.SetStaticPath("/static", tmpDir)
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/static/test.txt")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, fileContent, string(body))
	})
}
