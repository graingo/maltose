package mhttp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
)

const (
	DefaultServerName  = "default"
	defaultPort        = "8080"
	defaultOpenapiPath = "/api.json"
	defaultSwaggerPath = "/swagger"
)

// Server HTTP server structure.
type Server struct {
	RouterGroup
	engine       *gin.Engine
	config       *Config
	routes       []Route
	openapi      *openapi3.T
	preBindItems []preBindItem
	uni          *ut.UniversalTranslator
	translator   ut.Translator
	srv          *http.Server
	panicHandler func(r *Request, err error)
}

// New creates a new HTTP server.
func New(config ...*Config) *Server {
	conf := defaultConfig()
	if len(config) > 0 {
		conf = config[0]
	}

	// disable gin's default log output
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	// set to production mode
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	s := &Server{
		engine:       engine,
		config:       conf,
		preBindItems: make([]preBindItem, 0),
		panicHandler: func(r *Request, err error) {
			code := merror.Code(err)
			if code == mcode.CodeNil {
				r.String(500, fmt.Sprintf("Error: %s", err.Error()))
			} else {
				r.String(codeToHTTPStatus(code), code.Message())
			}
		},
	}

	// initialize root RouterGroup
	s.RouterGroup = RouterGroup{
		server:      s,
		path:        "/",
		ginGroup:    &s.engine.RouterGroup,
		middlewares: make([]MiddlewareFunc, 0),
		parent:      nil,
	}
	gin.Recovery()

	// add default middlewares
	s.Use(
		internalMiddlewareTrace(),
		internalMiddlewareRecovery(),
		internalMiddlewareMetric(),
		internalMiddlewareDefaultResponse(),
	)

	if s.config.ServerLocale != "" {
		// register translator
		s.registerValidateTranslator(s.config.ServerLocale)
	}

	return s
}

// WithPanicHandler sets a custom panic handler for the server
func (s *Server) WithPanicHandler(handler func(r *Request, err error)) *Server {
	s.panicHandler = handler
	return s
}
