package mhttp_test

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/examples/raw-stream/api"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/require"
)

type bufferedStreamWriter struct {
	gin.ResponseWriter
	buffer interface {
		io.Writer
		Flush() error
	}
}

func (w *bufferedStreamWriter) Write(p []byte) (int, error) { return w.buffer.Write(p) }
func (w *bufferedStreamWriter) FlushError() error {
	if err := w.buffer.Flush(); err != nil {
		return err
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}
func (w *bufferedStreamWriter) Flush() { _ = w.FlushError() }

func TestRawDownloadFlushesMiddlewareBuffers(t *testing.T) {
	for _, protocol := range []string{"HTTP1", "HTTP2"} {
		for _, encoding := range []string{"buffer", "gzip"} {
			t.Run(protocol+"/"+encoding, func(t *testing.T) {
				reader, writer := io.Pipe()
				release := make(chan struct{})
				defer close(release)
				s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
					return &api.DownloadRes{Body: reader}, nil
				}}, mhttp.MiddlewareLog(), func(r *mhttp.Request) {
					w := &bufferedStreamWriter{ResponseWriter: r.Writer, buffer: bufio.NewWriterSize(r.Writer, 4096)}
					if encoding == "gzip" {
						compressed := gzip.NewWriter(r.Writer)
						w.buffer = compressed
						r.Header("Content-Encoding", "gzip")
						defer compressed.Close()
					}
					r.Writer = w
					r.Next()
				})
				server := httptest.NewUnstartedServer(s)
				server.EnableHTTP2 = protocol == "HTTP2"
				server.StartTLS()
				defer server.Close()
				go func() {
					defer writer.Close()
					_, _ = writer.Write([]byte{0, 255})
					<-release
					_, _ = writer.Write([]byte{1})
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				req, err := http.NewRequestWithContext(ctx, "GET", server.URL+filePath, nil)
				require.NoError(t, err)
				res, err := server.Client().Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, protocol == "HTTP2", res.ProtoMajor == 2)
				first := make([]byte, 2)
				_, err = io.ReadFull(res.Body, first)
				require.NoError(t, err, "the first chunk must arrive while the source is still open")
				require.Equal(t, []byte{0, 255}, first)
				release <- struct{}{}
				rest, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, []byte{1}, rest)
			})
		}
	}
}

type flushFailureWriter struct {
	*httptest.ResponseRecorder
	failure error
	flushes int
}

func (w *flushFailureWriter) FlushError() error { w.flushes++; return w.failure }

type middlewareFlushFailure struct {
	gin.ResponseWriter
	failure error
}

func (w *middlewareFlushFailure) FlushError() error { return w.failure }
func (w *middlewareFlushFailure) Flush()            { _ = w.FlushError() }

func TestRawFlushFailureRemainsObservable(t *testing.T) {
	for _, name := range []string{"transport", "logging and buffering", "middleware"} {
		t.Run(name, func(t *testing.T) {
			failure := errors.New("transport flush failed")
			source := &streamBody{Reader: strings.NewReader("bytes")}
			observed := make(chan error, 1)
			middleware := []mhttp.MiddlewareFunc{observeError(observed), appStreamErrors}
			if name == "logging and buffering" {
				middleware = append(middleware, mhttp.MiddlewareLog(), func(r *mhttp.Request) {
					r.Writer = &bufferedStreamWriter{ResponseWriter: r.Writer, buffer: bufio.NewWriter(r.Writer)}
					r.Next()
				})
			}
			if name == "middleware" {
				middleware = append(middleware, func(r *mhttp.Request) {
					r.Writer = &middlewareFlushFailure{ResponseWriter: r.Writer, failure: failure}
					r.Next()
				}, mhttp.MiddlewareLog())
			}
			s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
				return &api.DownloadRes{Body: source}, nil
			}}, middleware...)
			w := &flushFailureWriter{ResponseRecorder: httptest.NewRecorder(), failure: failure}
			require.PanicsWithValue(t, http.ErrAbortHandler, func() { s.ServeHTTP(w, httptest.NewRequest("GET", filePath, nil)) })
			var stream *mhttp.StreamError
			recorded := <-observed
			require.ErrorAs(t, recorded, &stream)
			require.ErrorIs(t, recorded, failure)
			require.Equal(t, int64(5), stream.Written)
			require.Equal(t, "bytes", w.Body.String())
			wantFlushes := 1
			if name == "middleware" {
				wantFlushes = 0
			}
			require.Equal(t, wantFlushes, w.flushes, "each chunk flushes each reached layer once")
			require.Equal(t, int32(1), source.closes.Load())
		})
	}
}

func TestRawPrecommitPanicRestoresHeaders(t *testing.T) {
	for _, stage := range []string{"controller", "source"} {
		for _, encoding := range []string{"identity", "gzip"} {
			t.Run(stage+"/"+encoding, func(t *testing.T) {
				source := &streamBody{Reader: panicReader{}}
				s := streamServer(t, &fileController{download: func(ctx context.Context, _ *api.DownloadReq) (*api.DownloadRes, error) {
					r := mhttp.RequestFromCtx(ctx)
					r.Header("Content-Length", "9999")
					r.Header("Content-Type", "application/octet-stream")
					r.Header("Content-Disposition", `attachment; filename="file.bin"`)
					if encoding == "gzip" {
						r.Header("Content-Encoding", encoding)
					}
					if stage == "controller" {
						panic("controller failed")
					}
					return &api.DownloadRes{Body: source}, nil
				}}, func(r *mhttp.Request) { r.Header("X-Request-ID", "request"); r.Next() })
				s.WithPanicHandler(func(r *mhttp.Request, _ error) {
					r.Header("X-Error", "internal")
					r.JSON(500, map[string]string{"error": "internal"})
				})
				server := httptest.NewServer(s)
				defer server.Close()
				res, err := server.Client().Get(server.URL + filePath)
				require.NoError(t, err)
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, 500, res.StatusCode)
				require.JSONEq(t, `{"error":"internal"}`, string(body))
				require.Equal(t, "request", res.Header.Get("X-Request-ID"))
				require.Equal(t, "internal", res.Header.Get("X-Error"))
				require.Empty(t, res.Header.Get("Content-Disposition"))
				require.Empty(t, res.Header.Get("Content-Encoding"))
				require.Contains(t, res.Header.Get("Content-Type"), "application/json")
				if stage == "source" {
					require.Equal(t, int32(1), source.closes.Load())
				}
			})
		}
	}
}
