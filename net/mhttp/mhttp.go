package mhttp

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
)

var configureGinOnce sync.Once

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
	prepareOnce  sync.Once
	serverMu     sync.RWMutex
	panicHandler func(r *Request, err error)
}

// New creates a new HTTP server.
func New(config ...*Config) *Server {
	conf := defaultConfig()
	if len(config) > 0 && config[0] != nil {
		conf = cloneConfig(config[0])
	}

	configureGinOnce.Do(func() {
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard
		gin.SetMode(gin.ReleaseMode)
	})

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

	// Initialize the root router group.
	s.RouterGroup = RouterGroup{
		server:      s,
		path:        "/",
		ginGroup:    &s.engine.RouterGroup,
		middlewares: make([]MiddlewareFunc, 0),
		parent:      nil,
	}
	// Register framework middleware before user routes are bound.
	s.Use(
		internalMiddlewareTrace(),
		internalMiddlewareRecovery(),
		internalMiddlewareMetric(),
		internalMiddlewareDefaultResponse(),
	)

	if s.config.ServerLocale != "" {
		s.registerValidateTranslator(s.config.ServerLocale)
	}

	return s
}

// WithPanicHandler sets the handler used to convert recovered panics into responses.
func (s *Server) WithPanicHandler(handler func(r *Request, err error)) *Server {
	s.panicHandler = handler
	return s
}
