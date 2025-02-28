package mhttp

import (
	"time"

	"github.com/mingzaily/maltose/os/mlog"
	"github.com/spf13/cast"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Address        string
	ServerName     string
	ServerRoot     string
	ServerLocale   string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	OpenapiPath    string
	SwaggerPath    string

	SwaggerTemplate string
	Logger          *mlog.Logger
}

func NewConfig() ServerConfig {
	return ServerConfig{
		Address:        defaultPort,
		ServerName:     DefaultServerName,
		ServerLocale:   "zh",
		Logger:         mlog.New(),
		ReadTimeout:    time.Second * 60,
		WriteTimeout:   time.Second * 60,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20, // 1MB
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
	if v, ok := configMap["openapi_path"]; ok {
		s.config.OpenapiPath = cast.ToString(v)
	}
	if v, ok := configMap["swagger_path"]; ok {
		s.config.SwaggerPath = cast.ToString(v)
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
