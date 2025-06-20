package mhttp

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/spf13/viper"
)

// Config is the server configuration.
type Config struct {
	// basic config
	Address        string
	ServerName     string
	ServerRoot     string
	ServerLocale   string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int

	// Health check
	HealthCheck bool

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

func defaultConfig() Config {
	return Config{
		// basic config default values
		Address:        defaultPort,
		ServerName:     DefaultServerName,
		ServerLocale:   "zh",
		ReadTimeout:    time.Second * 60,
		WriteTimeout:   time.Second * 60,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20, // 1MB

		// Health check
		HealthCheck: true,

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
func (s *Server) SetConfigWithMap(configMap map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(configMap)
	return v.Unmarshal(&s.config)
}

// SetAddress sets the server listening address.
func (s *Server) SetAddress(addr string) {
	s.config.Address = addr
}

// SetServerName sets the server name.
func (s *Server) SetServerName(name string) {
	s.config.ServerName = name
}

// SetLogger sets the logger instance.
func (s *Server) SetLogger(logger *mlog.Logger) {
	s.config.Logger = logger
}

// Logger gets the logger instance.
func (s *Server) Logger() *mlog.Logger {
	return s.config.Logger
}

// SetConfigWithMap sets the server config.
func (s *Config) SetConfigWithMap(configMap map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(configMap)
	return v.Unmarshal(s)
}

// ConfigFromMap creates a new server config from a map.
func ConfigFromMap(configMap map[string]any) (Config, error) {
	config := defaultConfig()
	if err := config.SetConfigWithMap(configMap); err != nil {
		return Config{}, err
	}
	return config, nil
}
