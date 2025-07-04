package mhttp

import (
	"context"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/graingo/maltose/os/mlog"
)

// contextKey defines the context key type.
type contextKey string

const (
	// requestKey is the key for storing Request objects in the context.
	requestKey  contextKey = "MaltoseRequest"
	ResponseKey contextKey = "MaltoseResponse"
)

// Request is the request wrapper.
type Request struct {
	*gin.Context
	server *Server // server instance
}

// RequestFromCtx gets the Request object from the context.
func RequestFromCtx(ctx context.Context) *Request {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(requestKey); v != nil {
		if r, ok := v.(*Request); ok {
			return r
		}
	}
	return nil
}

func newRequest(c *gin.Context, s *Server) *Request {
	// try to get from context first
	if r := RequestFromCtx(c.Request.Context()); r != nil {
		return r
	}
	// create new Request object
	r := &Request{Context: c, server: s}
	// directly modify the original context, not create a new request
	r.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestKey, r))
	return r
}

// GetServerName gets the server name.
func (r *Request) GetServerName() string {
	return r.server.config.ServerName
}

// Logger gets the logger instance.
func (r *Request) Logger() *mlog.Logger {
	return r.server.logger()
}

// Conf gets the server config.
func (r *Request) Conf() *Config {
	return r.server.config
}

// GetHandlerResponse gets the handler response.
func (r *Request) GetHandlerResponse() any {
	res, _ := r.Get(string(ResponseKey))
	return res
}

// SetHandlerResponse sets the handler response.
func (r *Request) SetHandlerResponse(res any) {
	r.Set(string(ResponseKey), res)
}

// Error adds an error message.
func (r *Request) Error(err error) *Request {
	r.Errors = append(r.Errors, &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePrivate,
	})
	return r
}

// GetTranslator gets the translator.
func (r *Request) GetTranslator() ut.Translator {
	return r.server.translator
}
