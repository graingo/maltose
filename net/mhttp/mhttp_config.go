package mhttp

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/spf13/viper"
)

// Config is the server configuration.
type Config struct {
	// Address is the address of the server.
	Address string `mapstructure:"address"`
	// ServerName is the name of the server.
	ServerName string `mapstructure:"server_name"`
	// ServerRoot is the root directory of the server.
	ServerRoot string `mapstructure:"server_root"`
	// ServerLocale is the locale of the server.
	ServerLocale string `mapstructure:"server_locale"`
	// ReadTimeout is the timeout for reading the request.
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	// WriteTimeout is the timeout for writing the response.
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	// IdleTimeout is the timeout for idle connections.
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	// MaxHeaderBytes is the maximum number of bytes in the request header.
	MaxHeaderBytes int `mapstructure:"max_header_bytes"`
	// HealthCheck is the health check config.
	HealthCheck bool `mapstructure:"health_check"`
	// TLSEnable is the tls config.
	TLSEnable bool `mapstructure:"tls_enable"`
	// TLSCertFile is the path to the tls certificate file.
	TLSCertFile string `mapstructure:"tls_cert_file"`
	// TLSKeyFile is the path to the tls key file.
	TLSKeyFile string `mapstructure:"tls_key_file"`
	// TLSServerName is the server name for tls.
	TLSServerName string `mapstructure:"tls_server_name"`
	// GracefulEnable is the graceful shutdown config.
	GracefulEnable bool `mapstructure:"graceful_enable"`
	// GracefulTimeout is the timeout for graceful shutdown.
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout"`
	// GracefulWaitTime is the wait time for graceful shutdown.
	GracefulWaitTime time.Duration `mapstructure:"graceful_wait_time"`
	// OpenapiPath is the path to the openapi file.
	OpenapiPath string `mapstructure:"openapi_path"`
	// SwaggerPath is the path to the swagger file.
	SwaggerPath string `mapstructure:"swagger_path"`
	// SwaggerTemplate is the template for the swagger file.
	SwaggerTemplate string `mapstructure:"swagger_template"`
	// PrintRoutes is the print routes config.
	PrintRoutes bool `mapstructure:"print_routes"`
	// Logger is the logger config.
	Logger *mlog.Logger
}

func defaultConfig() *Config {
	return &Config{
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

		// PrintRoutes
		PrintRoutes: false,
	}
}

// SetConfigWithMap sets the server config.
func (s *Config) SetConfigWithMap(configMap map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(configMap)
	return v.Unmarshal(s)
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
func (s *Server) SetConfigWithMap(configMap map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(configMap)
	return v.Unmarshal(&s.config)
}

func (s *Server) SetConfig(config *Config) {
	s.config = config
}

// ConfigFromMap creates a new server config from a map.
func ConfigFromMap(configMap map[string]any) (*Config, error) {
	config := defaultConfig()
	if err := config.SetConfigWithMap(configMap); err != nil {
		return nil, err
	}
	return config, nil
}
