// Run from the repository root with RAW_STREAM_TOKEN set and generated documents:
// go run ./examples/raw-stream
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/graingo/maltose/examples/raw-stream/api"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/net/mhttp/contract"
)

type controller struct{ directory string }

func (c *controller) Upload(ctx context.Context, req *api.UploadReq) (*api.UploadRes, error) {
	file, err := os.CreateTemp(c.directory, "upload-")
	if err != nil {
		return nil, err
	}
	name := file.Name()
	defer os.Remove(name)
	n, copyErr := io.Copy(file, req.Body)
	closeErr := file.Close()
	if err = errors.Join(copyErr, closeErr, ctx.Err()); err != nil {
		return nil, err
	}
	if err = os.Rename(name, filepath.Join(c.directory, req.ID)); err != nil {
		return nil, err
	}
	return &api.UploadRes{Size: n}, nil
}
func (c *controller) Download(ctx context.Context, req *api.DownloadReq) (*api.DownloadRes, error) {
	file, err := os.Open(filepath.Join(c.directory, req.ID))
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	r := mhttp.RequestFromCtx(ctx)
	r.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	r.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.bin"`, req.ID))
	// Ownership transfers to Maltose on a successful return.
	return &api.DownloadRes{Body: file}, nil
}

func main() {
	token := os.Getenv("RAW_STREAM_TOKEN")
	if token == "" {
		log.Fatal("set RAW_STREAM_TOKEN for the example's bearer authentication")
	}
	dir, err := os.MkdirTemp("", "maltose-stream-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	s := mhttp.New()
	address := os.Getenv("RAW_STREAM_ADDRESS")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	s.SetAddress(address)
	s.Use(mhttp.MiddlewareLog(), func(r *mhttp.Request) {
		if r.GetHeader("Authorization") != "Bearer "+token {
			r.AbortWithStatusJSON(401, map[string]string{"error": "Unauthorized"})
			return
		}
		r.Next()
	}, applicationErrors, streamLimits)
	s.Bind(&controller{directory: dir})
	if err = s.LoadOpenAPIFile("examples/raw-stream/openapi.yaml"); err != nil {
		log.Fatal(err)
	}
	if err = s.Prepare(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err = s.ValidateOpenAPIRoutes(); err != nil {
		log.Fatal(err)
	}
	s.Run()
}

func streamLimits(r *mhttp.Request) {
	// Each application's upload intent determines its own limit and deadline.
	ctx, cancel := context.WithTimeout(r.Request.Context(), 30*time.Second)
	defer cancel()
	r.Request = r.Request.WithContext(ctx)
	transport := r.ResponseController()
	deadline, _ := ctx.Deadline()
	if err := transport.SetReadDeadline(deadline); err != nil {
		r.Error(err)
		r.Abort()
		return
	}
	if err := transport.SetWriteDeadline(deadline); err != nil {
		r.Error(err)
		r.Abort()
		return
	}
	// Cancellation interrupts blocked network I/O; context alone is a signal.
	completed := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(completed)
		_ = transport.SetReadDeadline(time.Now())
		_ = transport.SetWriteDeadline(time.Now())
	})
	defer func() {
		if !stop() {
			<-completed
		}
	}()
	if r.RawRequestMediaType() != "" {
		r.Request.Body = http.MaxBytesReader(r.Writer, r.Request.Body, 32<<20)
	}
	r.Next()
}

func applicationErrors(r *mhttp.Request) {
	r.Next()
	if r.Writer.Written() || len(r.Errors) == 0 {
		return
	}
	err := r.Errors.Last().Err
	status := http.StatusInternalServerError
	var size *http.MaxBytesError
	var decode *contract.DecodeError
	var validation *contract.ValidationError
	var timeout interface{ Timeout() bool }
	switch {
	case errors.As(err, &size):
		status = http.StatusRequestEntityTooLarge
	case errors.As(err, &decode):
		status = decode.Status
	case errors.As(err, &validation):
		status = http.StatusBadRequest
	case errors.Is(err, os.ErrNotExist):
		status = http.StatusNotFound
	case errors.Is(err, context.DeadlineExceeded) || errors.As(err, &timeout) && timeout.Timeout():
		status = http.StatusRequestTimeout
	}
	r.JSON(status, map[string]string{"error": http.StatusText(status)})
}
