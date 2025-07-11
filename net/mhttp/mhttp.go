package mhttp

import (
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

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
	}

	// initialize root RouterGroup
	s.RouterGroup = RouterGroup{
		server:      s,
		path:        "/",
		ginGroup:    &s.engine.RouterGroup,
		middlewares: make([]MiddlewareFunc, 0),
		parent:      nil,
	}

	// add default middlewares
	s.Use(
		internalMiddlewareRecovery(),
		internalMiddlewareTrace(),
		internalMiddlewareMetric(),
		internalMiddlewareDefaultResponse(),
	)

	if s.config.ServerLocale != "" {
		// register translator
		s.registerValidateTranslator(s.config.ServerLocale)
	}

	return s
}
