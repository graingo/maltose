package mhttp

import (
	"time"

	"github.com/savorelle/maltose/os/mlog"
	"github.com/spf13/cast"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	// 基础配置
	Address        string
	ServerName     string
	ServerRoot     string
	ServerLocale   string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int

	// TLS 配置
	TLSEnable     bool
	TLSCertFile   string
	TLSKeyFile    string
	TLSAutoEnable bool
	TLSServerName string

	// 优雅关闭配置
	GracefulEnable   bool
	GracefulTimeout  time.Duration
	GracefulWaitTime time.Duration

	// API 文档配置
	OpenapiPath     string
	SwaggerPath     string
	SwaggerTemplate string

	// 日志配置
	Logger *mlog.Logger
}

func NewConfig() ServerConfig {
	return ServerConfig{
		// 基础配置默认值
		Address:        defaultPort,
		ServerName:     DefaultServerName,
		ServerLocale:   "zh",
		ReadTimeout:    time.Second * 60,
		WriteTimeout:   time.Second * 60,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20, // 1MB

		// TLS 默认配置
		TLSEnable:     false,
		TLSAutoEnable: false,

		// 优雅关闭默认配置
		GracefulEnable:   true,
		GracefulTimeout:  time.Second * 30,
		GracefulWaitTime: time.Second * 5,

		// 日志默认配置
		Logger: mlog.New(),
	}
}

// SetConfig 设置服务器配置
func (s *Server) SetConfig(configMap map[string]any) {
	if v, ok := configMap["address"]; ok {
		s.config.Address = cast.ToString(v)
	}
	if v, ok := configMap["server_name"]; ok {
		s.config.ServerName = cast.ToString(v)
	}
	if v, ok := configMap["server_root"]; ok {
		s.config.ServerRoot = cast.ToString(v)
	}
	if v, ok := configMap["server_locale"]; ok {
		s.config.ServerLocale = cast.ToString(v)
	}
	if v, ok := configMap["read_timeout"]; ok {
		s.config.ReadTimeout = cast.ToDuration(v)
	}
	if v, ok := configMap["write_timeout"]; ok {
		s.config.WriteTimeout = cast.ToDuration(v)
	}
	if v, ok := configMap["idle_timeout"]; ok {
		s.config.IdleTimeout = cast.ToDuration(v)
	}
	if v, ok := configMap["max_header_bytes"]; ok {
		s.config.MaxHeaderBytes = cast.ToInt(v)
	}

	// TLS 配置
	if v, ok := configMap["tls_enable"]; ok {
		s.config.TLSEnable = cast.ToBool(v)
	}
	if v, ok := configMap["tls_cert_file"]; ok {
		s.config.TLSCertFile = cast.ToString(v)
	}
	if v, ok := configMap["tls_key_file"]; ok {
		s.config.TLSKeyFile = cast.ToString(v)
	}
	if v, ok := configMap["tls_auto_enable"]; ok {
		s.config.TLSAutoEnable = cast.ToBool(v)
	}
	if v, ok := configMap["tls_server_name"]; ok {
		s.config.TLSServerName = cast.ToString(v)
	}

	// 优雅关闭配置
	if v, ok := configMap["graceful_enable"]; ok {
		s.config.GracefulEnable = cast.ToBool(v)
	}
	if v, ok := configMap["graceful_timeout"]; ok {
		s.config.GracefulTimeout = cast.ToDuration(v)
	}
	if v, ok := configMap["graceful_wait_time"]; ok {
		s.config.GracefulWaitTime = cast.ToDuration(v)
	}

	// API 文档配置
	if v, ok := configMap["openapi_path"]; ok {
		s.config.OpenapiPath = cast.ToString(v)
	}
	if v, ok := configMap["swagger_path"]; ok {
		s.config.SwaggerPath = cast.ToString(v)
	}
	if v, ok := configMap["swagger_template"]; ok {
		s.config.SwaggerTemplate = cast.ToString(v)
	}
}

// SetAddress 设置服务器监听地址
func (s *Server) SetAddress(addr string) {
	s.config.Address = addr
}

// SetServerName
func (s *Server) SetServerName(name string) {
	s.config.ServerName = name
}

// Logger 获取日志实例
func (s *Server) Logger() *mlog.Logger {
	return s.config.Logger
}
