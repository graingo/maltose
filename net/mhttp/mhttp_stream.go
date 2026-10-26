package mhttp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// StreamError reports a failure after a raw success response has started.
// The connection/HTTP stream is aborted after middleware has observed this error.
type StreamError struct {
	Written int64
	Cause   error
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("raw response failed after %d bytes: %v", e.Written, e.Cause)
}
func (e *StreamError) Unwrap() error { return e.Cause }

// RawRequestMediaType returns the declared raw request type, or an empty string.
func (r *Request) RawRequestMediaType() string {
	if r.operation == nil {
		return ""
	}
	return r.operation.RawRequestMedia
}

// RawResponseMediaType returns the declared raw success type, or an empty string.
func (r *Request) RawResponseMediaType() string {
	if r.operation == nil {
		return ""
	}
	return r.operation.RawResponseMedia
}

// ResponseController exposes the underlying HTTP transport's deadline and duplex
// controls, including when Gin or logging wrappers surround its response writer.
func (r *Request) ResponseController() *http.ResponseController {
	writer, ok := r.Request.Context().Value(transportWriterKey).(http.ResponseWriter)
	if !ok {
		writer = r.Writer
	}
	return http.NewResponseController(writer)
}

const transportWriterKey contextKey = "MaltoseTransportWriter"

// streamFlusher sits below application writers and preserves transport errors
// across Gin's void Flush method. Writes and flushes follow the same
// middleware chain; deadline and duplex controls target the transport directly.
type streamFlusher struct {
	gin.ResponseWriter
	transport *http.ResponseController
	err       error
}

func (w *streamFlusher) FlushError() error {
	w.WriteHeaderNow()
	w.err = w.transport.Flush()
	return w.err
}
func (w *streamFlusher) Flush()                      { _ = w.FlushError() }
func (w *streamFlusher) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *streamFlusher) flush(writer gin.ResponseWriter) error {
	err := http.NewResponseController(writer).Flush()
	if err == nil {
		err = w.err
	}
	if errors.Is(err, http.ErrNotSupported) {
		return nil
	}
	return err
}

type streamResponse struct {
	body     io.ReadCloser
	reader   *bufio.Reader
	once     sync.Once
	closeErr error
	stop     func() bool
}

func (s *streamResponse) close() error {
	s.once.Do(func() { s.closeErr = s.body.Close() })
	return s.closeErr
}
func (s *streamResponse) finish() error {
	if s.stop != nil {
		s.stop()
	}
	return s.close()
}

type streamReader struct {
	ctx    context.Context
	source io.Reader
}

func (s streamReader) Read(p []byte) (int, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := s.source.Read(p)
	if canceled := s.ctx.Err(); canceled != nil {
		return n, canceled
	}
	return n, err
}
func (r *Request) prepareStreamResponse(response any) error {
	node := r.operation.Types[r.operation.Response]
	field := node.Fields[0]
	body, ok := reflect.ValueOf(response).Elem().FieldByIndex(field.Index).Interface().(io.ReadCloser)
	if !ok || isNilStream(body) {
		return fmt.Errorf("raw success response requires a non-nil io.ReadCloser")
	}
	stream := &streamResponse{body: body}
	r.stream = stream
	ctx := r.Request.Context()
	stream.stop = context.AfterFunc(ctx, func() { _ = stream.close() })
	if r.RawRequestMediaType() != "" && r.Request.ProtoMajor == 1 {
		// HTTP/1 otherwise drains the unread request before starting the response.
		// HTTP/2 supports duplex without this HTTP/1 control.
		if err := r.ResponseController().EnableFullDuplex(); err != nil {
			return errors.Join(err, stream.finish())
		}
	}
	stream.reader = bufio.NewReaderSize(streamReader{ctx: ctx, source: body}, 32*1024)
	// A source that fails before producing a byte goes through application errors.
	if _, err := stream.reader.Peek(1); err != nil && err != io.EOF {
		return errors.Join(err, stream.finish())
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, stream.finish())
	}
	return nil
}
func isNilStream(body io.ReadCloser) bool {
	if body == nil {
		return true
	}
	v := reflect.ValueOf(body)
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan, reflect.Interface:
		return v.IsNil()
	}
	return false
}

type streamWriter struct{ request *Request }

func (w streamWriter) Write(p []byte) (int, error) {
	n, err := w.request.Writer.Write(p)
	if err == nil {
		err = w.request.streamFlush.flush(w.request.Writer)
	}
	return n, err
}
func (r *Request) writeStreamResponse() {
	if r.stream == nil {
		r.Error(fmt.Errorf("raw response requires a bound controller response"))
		return
	}
	r.Header("Content-Type", r.RawResponseMediaType())
	r.Status(r.SuccessStatus())
	r.Writer.WriteHeaderNow()
	written, err := io.CopyBuffer(streamWriter{r}, r.stream.reader, make([]byte, 32*1024))
	if length := r.Writer.Header().Get("Content-Length"); err == nil && length != "" {
		expected, parseErr := strconv.ParseInt(length, 10, 64)
		if parseErr != nil || expected != written {
			err = io.ErrUnexpectedEOF
		}
	}
	err = errors.Join(err, r.stream.finish())
	if err != nil {
		r.Error(&StreamError{Written: written, Cause: err})
		r.abortStream = true
	}
}
