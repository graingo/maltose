package mhttp

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/graingo/mconv"
)

// ServerConfig is the server configuration.
type ServerConfig struct {
	// basic config
	Address        string
	ServerName     string
	ServerRoot     string
	ServerLocale   string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int

	// TLS config
	TLSEnable     bool
	TLSCertFile   string
	TLSKeyFile    string
	TLSServerName string

	// graceful shutdown config
	GracefulEnable   bool
	GracefulTimeout  time.Duration
	GracefulWaitTime time.Duration

	// API doc config
	OpenapiPath     string
	SwaggerPath     string
	SwaggerTemplate string

	// log config
	Logger *mlog.Logger
}

func NewConfig() ServerConfig {
	return ServerConfig{
		// basic config default values
		Address:        defaultPort,
		ServerName:     DefaultServerName,
		ServerLocale:   "zh",
		ReadTimeout:    time.Second * 60,
		WriteTimeout:   time.Second * 60,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20, // 1MB

		// TLS default config
		TLSEnable: false,

		// graceful shutdown default config
		GracefulEnable:   true,
		GracefulTimeout:  time.Second * 30,
		GracefulWaitTime: time.Second * 5,

		// log default config
		Logger: mlog.New(),
	}
}

// SetConfigWithMap sets the server config.
func (s *Server) SetConfigWithMap(configMap map[string]any) {
	if v, ok := configMap["address"]; ok {
		s.config.Address = mconv.ToString(v)
	}
	if v, ok := configMap["server_name"]; ok {
		s.config.ServerName = mconv.ToString(v)
	}
	if v, ok := configMap["server_root"]; ok {
		s.config.ServerRoot = mconv.ToString(v)
	}
	if v, ok := configMap["server_locale"]; ok {
		s.config.ServerLocale = mconv.ToString(v)
	}
	if v, ok := configMap["read_timeout"]; ok {
		s.config.ReadTimeout = mconv.ToDuration(v)
	}
	if v, ok := configMap["write_timeout"]; ok {
		s.config.WriteTimeout = mconv.ToDuration(v)
	}
	if v, ok := configMap["idle_timeout"]; ok {
		s.config.IdleTimeout = mconv.ToDuration(v)
	}
	if v, ok := configMap["max_header_bytes"]; ok {
		s.config.MaxHeaderBytes = mconv.ToInt(v)
	}

	// TLS config
	if v, ok := configMap["tls_enable"]; ok {
		s.config.TLSEnable = mconv.ToBool(v)
	}
	if v, ok := configMap["tls_cert_file"]; ok {
		s.config.TLSCertFile = mconv.ToString(v)
	}
	if v, ok := configMap["tls_key_file"]; ok {
		s.config.TLSKeyFile = mconv.ToString(v)
	}
	if v, ok := configMap["tls_server_name"]; ok {
		s.config.TLSServerName = mconv.ToString(v)
	}

	// graceful shutdown config
	if v, ok := configMap["graceful_enable"]; ok {
		s.config.GracefulEnable = mconv.ToBool(v)
	}
	if v, ok := configMap["graceful_timeout"]; ok {
		s.config.GracefulTimeout = mconv.ToDuration(v)
	}
	if v, ok := configMap["graceful_wait_time"]; ok {
		s.config.GracefulWaitTime = mconv.ToDuration(v)
	}

	// API doc config
	if v, ok := configMap["openapi_path"]; ok {
		s.config.OpenapiPath = mconv.ToString(v)
	}
	if v, ok := configMap["swagger_path"]; ok {
		s.config.SwaggerPath = mconv.ToString(v)
	}
	if v, ok := configMap["swagger_template"]; ok {
		s.config.SwaggerTemplate = mconv.ToString(v)
	}
}

// SetAddress sets the server listening address.
func (s *Server) SetAddress(addr string) {
	s.config.Address = addr
}

// SetServerName sets the server name.
func (s *Server) SetServerName(name string) {
	s.config.ServerName = name
}

// Logger gets the logger instance.
func (s *Server) Logger() *mlog.Logger {
	return s.config.Logger
}
