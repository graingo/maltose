package mhttp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/graingo/maltose/examples/raw-stream/api"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

const fileID = "ccba9e91-bfa0-468e-91ac-956ee9097271"
const filePath = "/files/" + fileID

type fileController struct {
	upload   func(context.Context, *api.UploadReq) (*api.UploadRes, error)
	download func(context.Context, *api.DownloadReq) (*api.DownloadRes, error)
}

func (c *fileController) Upload(ctx context.Context, r *api.UploadReq) (*api.UploadRes, error) {
	return c.upload(ctx, r)
}
func (c *fileController) Download(ctx context.Context, r *api.DownloadReq) (*api.DownloadRes, error) {
	return c.download(ctx, r)
}
func streamServer(t *testing.T, c *fileController, m ...mhttp.MiddlewareFunc) *mhttp.Server {
	t.Helper()
	s := mhttp.New()
	s.Use(m...)
	s.Bind(c)
	upload, err := contract.Compile(contract.TypeOf[api.UploadReq](), contract.TypeOf[api.UploadRes]())
	require.NoError(t, err)
	download, err := contract.Compile(contract.TypeOf[api.DownloadReq](), contract.TypeOf[api.DownloadRes]())
	require.NoError(t, err)
	artifact, err := contract.Generate([]*contract.Operation{upload, download}, contract.Options{}, nil)
	require.NoError(t, err)
	require.NoError(t, s.LoadOpenAPI(artifact.Document, artifact.Manifest))
	require.NoError(t, s.Prepare(context.Background()))
	require.NoError(t, s.ValidateOpenAPIRoutes())
	return s
}

type streamBody struct {
	io.Reader
	reads, closes atomic.Int32
}

func (b *streamBody) Read(p []byte) (int, error) { b.reads.Add(1); return b.Reader.Read(p) }
func (b *streamBody) Close() error               { b.closes.Add(1); return nil }

func TestRawAuthenticationAndLoggingBeforeRead(t *testing.T) {
	previous := mhttp.LogMaxBodySize
	mhttp.LogMaxBodySize = -1
	logMiddleware := mhttp.MiddlewareLog()
	mhttp.LogMaxBodySize = previous
	body := &streamBody{Reader: strings.NewReader("secret bytes")}
	called := false
	s := streamServer(t, &fileController{upload: func(_ context.Context, r *api.UploadReq) (*api.UploadRes, error) {
		called = true
		require.Same(t, body, r.Body)
		require.Zero(t, body.reads.Load())
		n, err := io.Copy(io.Discard, r.Body)
		return &api.UploadRes{Size: n}, err
	}}, logMiddleware, func(r *mhttp.Request) {
		require.Equal(t, "application/octet-stream", r.RawRequestMediaType())
		require.Zero(t, body.reads.Load())
		if r.GetHeader("Authorization") == "" {
			r.AbortWithStatus(401)
			return
		}
		r.Next()
	}, mhttp.MiddlewareResponse())
	for _, tc := range []struct {
		auth, media, path string
		want              int
	}{
		{"", "application/octet-stream", filePath, 401},
		{"token", "application/json", filePath, 415},
		{"token", "application/octet-stream", "/files/bad", 400},
		{"token", "application/octet-stream", filePath, 200},
	} {
		r := httptest.NewRequest("PUT", tc.path, body)
		r.Header.Set("Content-Type", tc.media)
		r.Header.Set("Authorization", tc.auth)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		require.Equal(t, tc.want, w.Code)
		if tc.want != 200 {
			require.False(t, called)
			require.Zero(t, body.reads.Load())
		} else {
			require.JSONEq(t, `{"size":12}`, w.Body.String())
		}
		require.Zero(t, body.closes.Load(), "incoming body remains owned by net/http")
	}
}

func TestRawBinaryHTTPAndChunkedUpload(t *testing.T) {
	raw := bytes.Repeat([]byte{0, 255, 254, 1, 2}, 1<<20)
	source := &streamBody{Reader: bytes.NewReader(raw)}
	uploaded := make(chan []byte, 1)
	framing := make(chan []string, 1)
	s := streamServer(t, &fileController{
		upload: func(ctx context.Context, r *api.UploadReq) (*api.UploadRes, error) {
			framing <- mhttp.RequestFromCtx(ctx).Request.TransferEncoding
			value, err := io.ReadAll(r.Body)
			uploaded <- value
			return &api.UploadRes{Size: int64(len(value))}, err
		},
		download: func(ctx context.Context, _ *api.DownloadReq) (*api.DownloadRes, error) {
			r := mhttp.RequestFromCtx(ctx)
			r.Header("Content-Disposition", `attachment; filename="sample.bin"`)
			return &api.DownloadRes{Body: source}, nil
		},
	}, mhttp.MiddlewareResponse())
	server := httptest.NewServer(s)
	defer server.Close()
	// Hide the byte reader's known length so the client sends chunked encoding.
	request, err := http.NewRequest("PUT", server.URL+filePath, io.NopCloser(bytes.NewReader(raw)))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/octet-stream")
	res, err := server.Client().Do(request)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, raw, <-uploaded)
	require.Equal(t, []string{"chunked"}, <-framing)
	result, err := server.Client().Get(server.URL + filePath)
	require.NoError(t, err)
	defer result.Body.Close()
	require.Equal(t, "application/octet-stream", result.Header.Get("Content-Type"))
	require.Contains(t, result.Header.Get("Content-Disposition"), "sample.bin")
	got, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, raw, got)
	require.Equal(t, int32(1), source.closes.Load())
}

func TestRawDownloadFlushesBeforeSourceCompletes(t *testing.T) {
	reader, writer := io.Pipe()
	source := &streamBody{Reader: reader}
	release := make(chan struct{})
	defer close(release)
	s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
		return &api.DownloadRes{Body: source}, nil
	}})
	server := httptest.NewServer(s)
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
	result, err := server.Client().Do(req)
	require.NoError(t, err)
	defer result.Body.Close()
	first := make([]byte, 2)
	_, err = io.ReadFull(result.Body, first)
	require.NoError(t, err)
	require.Equal(t, []byte{0, 255}, first)
	release <- struct{}{}
	rest, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, []byte{1}, rest)
	require.Equal(t, int32(1), source.closes.Load())
}

type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }
func observeError(ch chan<- error) mhttp.MiddlewareFunc {
	return func(r *mhttp.Request) {
		r.Next()
		if e := r.Errors.Last(); e != nil {
			ch <- e.Err
		}
	}
}
func appStreamErrors(r *mhttp.Request) {
	r.Next()
	if r.Writer.Written() || len(r.Errors) == 0 {
		return
	}
	err := r.Errors.Last().Err
	status := 422
	var size *http.MaxBytesError
	if errors.As(err, &size) {
		status = 413
	}
	if errors.Is(err, context.DeadlineExceeded) {
		status = 408
	}
	r.JSON(status, map[string]any{"error": err.Error()})
}
func TestRawResponseFailureBoundary(t *testing.T) {
	failure := errors.New("source failed")
	for _, tc := range []struct {
		name   string
		prefix []byte
		want   int
	}{
		{"before", nil, 422}, {"after", []byte{0, 255, 254}, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &streamBody{Reader: io.MultiReader(bytes.NewReader(tc.prefix), failingReader{failure})}
			observed := make(chan error, 1)
			s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
				return &api.DownloadRes{Body: source}, nil
			}}, observeError(observed), appStreamErrors)
			server := httptest.NewServer(s)
			defer server.Close()
			res, err := server.Client().Get(server.URL + filePath)
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, tc.want, res.StatusCode)
			got, err := io.ReadAll(res.Body)
			recorded := <-observed
			require.ErrorIs(t, recorded, failure)
			if tc.want == 200 {
				require.ErrorIs(t, err, io.ErrUnexpectedEOF)
				require.Equal(t, tc.prefix, got)
				var stream *mhttp.StreamError
				require.ErrorAs(t, recorded, &stream)
				require.Equal(t, int64(len(tc.prefix)), stream.Written)
			} else {
				require.NoError(t, err)
				require.JSONEq(t, `{"error":"source failed"}`, string(got))
			}
			require.Equal(t, int32(1), source.closes.Load())
		})
	}
}

func TestRawSizeLimit(t *testing.T) {
	observed := make(chan error, 1)
	s := streamServer(t, &fileController{upload: func(_ context.Context, r *api.UploadReq) (*api.UploadRes, error) {
		n, err := io.Copy(io.Discard, r.Body)
		return &api.UploadRes{Size: n}, err
	}}, observeError(observed), appStreamErrors, func(r *mhttp.Request) {
		r.Request.Body = http.MaxBytesReader(r.Writer, r.Request.Body, 3)
		r.Next()
	})
	server := httptest.NewServer(s)
	defer server.Close()
	req, err := http.NewRequest("PUT", server.URL+filePath, io.NopCloser(strings.NewReader("12345")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := server.Client().Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 413, res.StatusCode)
	var limit *http.MaxBytesError
	require.ErrorAs(t, <-observed, &limit)
}

type blockingStream struct {
	started, closed chan struct{}
	once            sync.Once
	closes          atomic.Int32
}

func (b *blockingStream) Read([]byte) (int, error) {
	b.once.Do(func() { close(b.started) })
	<-b.closed
	return 0, io.EOF
}
func (b *blockingStream) Close() error {
	if b.closes.Add(1) == 1 {
		close(b.closed)
	}
	return nil
}
func TestRawCancellationClosesSource(t *testing.T) {
	source := &blockingStream{started: make(chan struct{}), closed: make(chan struct{})}
	observed := make(chan error, 1)
	s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
		return &api.DownloadRes{Body: source}, nil
	}}, observeError(observed), appStreamErrors)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	w := httptest.NewRecorder()
	go func() { defer close(done); s.ServeHTTP(w, httptest.NewRequest("GET", filePath, nil).WithContext(ctx)) }()
	select {
	case <-source.started:
	case <-time.After(3 * time.Second):
		t.Fatal("source not read")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("cancel did not unblock source")
	}
	require.ErrorIs(t, <-observed, context.Canceled)
	require.Equal(t, int32(1), source.closes.Load())
	require.Equal(t, 422, w.Code)
}

func TestRawReadDeadlineAndTotalTimeout(t *testing.T) {
	for _, total := range []bool{false, true} {
		t.Run(fmt.Sprint(total), func(t *testing.T) {
			observed := make(chan error, 1)
			s := streamServer(t, &fileController{upload: func(ctx context.Context, r *api.UploadReq) (*api.UploadRes, error) {
				n, err := io.Copy(io.Discard, r.Body)
				if total && ctx.Err() != nil {
					err = ctx.Err()
				}
				return &api.UploadRes{Size: n}, err
			}}, observeError(observed), appStreamErrors, func(r *mhttp.Request) {
				if total {
					ctx, cancel := context.WithTimeout(r.Request.Context(), 50*time.Millisecond)
					defer cancel()
					r.Request = r.Request.WithContext(ctx)
					stop := context.AfterFunc(ctx, func() { _ = r.ResponseController().SetReadDeadline(time.Now()) })
					defer stop()
				} else {
					require.NoError(t, r.ResponseController().SetReadDeadline(time.Now().Add(50*time.Millisecond)))
				}
				r.Next()
			})
			server := httptest.NewServer(s)
			defer server.Close()
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "PUT", server.URL+filePath, reader)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/octet-stream")
			res, err := server.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			recorded := <-observed
			if total {
				require.Equal(t, 408, res.StatusCode)
				require.ErrorIs(t, recorded, context.DeadlineExceeded)
			} else {
				require.Equal(t, 422, res.StatusCode)
				var timeout interface{ Timeout() bool }
				require.ErrorAs(t, recorded, &timeout)
				require.True(t, timeout.Timeout())
			}
		})
	}
}

type EchoReq struct {
	mmeta.Meta `method:"POST" path:"/echo" body:"raw" response:"raw" status:"201"`
	Body       io.ReadCloser `stream:"body" field:"required"`
}
type EchoRes struct {
	Body io.ReadCloser `stream:"body"`
}
type echoController struct{}

func (*echoController) Echo(_ context.Context, r *EchoReq) (*EchoRes, error) {
	return &EchoRes{Body: io.NopCloser(r.Body)}, nil
}
func TestRawDuplexOverHTTP(t *testing.T) {
	s := mhttp.New()
	s.Bind(&echoController{})
	require.NoError(t, s.Prepare(context.Background()))
	server := httptest.NewServer(s)
	defer server.Close()
	for _, size := range []int{0, 64 * 1024, 1024 * 1024} {
		data := bytes.Repeat([]byte{0, 255}, size/2)
		req, err := http.NewRequest("POST", server.URL+"/echo", io.NopCloser(bytes.NewReader(data)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/octet-stream")
		res, err := server.Client().Do(req)
		require.NoError(t, err)
		got, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)
		require.Equal(t, 201, res.StatusCode)
		require.True(t, bytes.Equal(data, got))
	}
}

type failingWriter struct {
	*httptest.ResponseRecorder
	failure error
}

func (w failingWriter) Write([]byte) (int, error) { return 0, w.failure }
func TestRawWriteFailure(t *testing.T) {
	failure := errors.New("transport failed")
	source := &streamBody{Reader: strings.NewReader("bytes")}
	observed := make(chan error, 1)
	s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
		return &api.DownloadRes{Body: source}, nil
	}}, observeError(observed), appStreamErrors)
	w := failingWriter{httptest.NewRecorder(), failure}
	require.PanicsWithValue(t, http.ErrAbortHandler, func() { s.ServeHTTP(w, httptest.NewRequest("GET", filePath, nil)) })
	err := <-observed
	require.ErrorIs(t, err, failure)
	var stream *mhttp.StreamError
	require.ErrorAs(t, err, &stream)
	require.Zero(t, stream.Written)
	require.Empty(t, w.Body.String())
	require.Equal(t, int32(1), source.closes.Load())
}

func TestRawNilSourceAndBusinessFailure(t *testing.T) {
	for _, kind := range []string{"nil", "typed nil", "business"} {
		t.Run(kind, func(t *testing.T) {
			s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
				switch kind {
				case "typed nil":
					var source *streamBody
					return &api.DownloadRes{Body: source}, nil
				case "business":
					return nil, errors.New("storage unavailable")
				default:
					return &api.DownloadRes{}, nil
				}
			}}, appStreamErrors)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, httptest.NewRequest("GET", filePath, nil))
			require.Equal(t, 422, w.Code)
			require.Contains(t, w.Header().Get("Content-Type"), "application/json")
		})
	}
}

func TestRawRuntimeRejectsStaleTransport(t *testing.T) {
	upload, err := contract.Compile(contract.TypeOf[api.UploadReq](), contract.TypeOf[api.UploadRes]())
	require.NoError(t, err)
	download, err := contract.Compile(contract.TypeOf[api.DownloadReq](), contract.TypeOf[api.DownloadRes]())
	require.NoError(t, err)
	download.RawResponseMedia = "image/png"
	artifact, err := contract.Generate([]*contract.Operation{upload, download}, contract.Options{}, nil)
	require.NoError(t, err)
	s := mhttp.New()
	s.Bind(&fileController{})
	require.NoError(t, s.LoadOpenAPI(artifact.Document, artifact.Manifest))
	require.ErrorContains(t, s.Prepare(context.Background()), "contract differs")
}

func TestRawFailureRestoresMiddlewareHeaders(t *testing.T) {
	source := &streamBody{Reader: failingReader{errors.New("source failed")}}
	s := streamServer(t, &fileController{download: func(ctx context.Context, _ *api.DownloadReq) (*api.DownloadRes, error) {
		r := mhttp.RequestFromCtx(ctx)
		r.Header("Content-Length", "9999")
		r.Header("Content-Type", "application/octet-stream")
		r.Header("Content-Encoding", "gzip")
		r.Header("Content-Disposition", `attachment; filename="file.bin"`)
		return &api.DownloadRes{Body: source}, nil
	}}, func(r *mhttp.Request) { r.Header("X-Request-ID", "request"); r.Next() }, appStreamErrors)
	server := httptest.NewServer(s)
	defer server.Close()
	res, err := server.Client().Get(server.URL + filePath)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, 422, res.StatusCode)
	require.JSONEq(t, `{"error":"source failed"}`, string(body))
	require.Equal(t, "request", res.Header.Get("X-Request-ID"))
	require.Empty(t, res.Header.Get("Content-Disposition"))
	require.Empty(t, res.Header.Get("Content-Encoding"))
	require.Contains(t, res.Header.Get("Content-Type"), "application/json")
	require.Equal(t, int32(1), source.closes.Load())
}

type closeFailureBody struct {
	streamBody
	failure error
}

func (b *closeFailureBody) Close() error { b.closes.Add(1); return b.failure }
func TestRawCloseAndLengthFailures(t *testing.T) {
	for _, kind := range []string{"close", "length"} {
		t.Run(kind, func(t *testing.T) {
			failure := errors.New("close failed")
			source := &closeFailureBody{streamBody: streamBody{Reader: strings.NewReader("abc")}}
			observed := make(chan error, 1)
			if kind == "close" {
				source.failure = failure
			}
			s := streamServer(t, &fileController{download: func(ctx context.Context, _ *api.DownloadReq) (*api.DownloadRes, error) {
				if kind == "length" {
					mhttp.RequestFromCtx(ctx).Header("Content-Length", "6")
				}
				return &api.DownloadRes{Body: source}, nil
			}}, observeError(observed))
			w := httptest.NewRecorder()
			require.PanicsWithValue(t, http.ErrAbortHandler, func() { s.ServeHTTP(w, httptest.NewRequest("GET", filePath, nil)) })
			err := <-observed
			var stream *mhttp.StreamError
			require.ErrorAs(t, err, &stream)
			require.Equal(t, int64(3), stream.Written)
			if kind == "close" {
				require.ErrorIs(t, err, failure)
			} else {
				require.ErrorIs(t, err, io.ErrUnexpectedEOF)
			}
			require.Equal(t, "abc", w.Body.String())
			require.Equal(t, int32(1), source.closes.Load())
		})
	}
}
func TestRawHTTP2AbortsFailedStream(t *testing.T) {
	failure := errors.New("source failed")
	source := &streamBody{Reader: io.MultiReader(strings.NewReader("abc"), failingReader{failure})}
	observed := make(chan error, 1)
	s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
		return &api.DownloadRes{Body: source}, nil
	}}, observeError(observed), appStreamErrors)
	server := httptest.NewUnstartedServer(s)
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	res, err := server.Client().Get(server.URL + filePath)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 2, res.ProtoMajor)
	body, err := io.ReadAll(res.Body)
	require.Error(t, err)
	require.Equal(t, "abc", string(body))
	require.ErrorIs(t, <-observed, failure)
	require.Equal(t, int32(1), source.closes.Load())
}

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) { panic("source panic") }
func TestRawSourcePanicKeepsCommittedBytes(t *testing.T) {
	source := &streamBody{Reader: io.MultiReader(strings.NewReader("abc"), panicReader{})}
	s := streamServer(t, &fileController{download: func(context.Context, *api.DownloadReq) (*api.DownloadRes, error) {
		return &api.DownloadRes{Body: source}, nil
	}})
	w := httptest.NewRecorder()
	require.PanicsWithValue(t, http.ErrAbortHandler, func() { s.ServeHTTP(w, httptest.NewRequest("GET", filePath, nil)) })
	require.Equal(t, "abc", w.Body.String())
	require.Equal(t, int32(1), source.closes.Load())
}
