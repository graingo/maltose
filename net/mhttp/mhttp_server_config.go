package mhttp

import (
	"time"

	"github.com/mingzaily/maltose/os/mlog"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Address    string
	ServerName string
	ServerRoot string
	Logger     *mlog.Logger

	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
}

func NewConfig() ServerConfig {
	return ServerConfig{
		Address:        defaultPort,
		ServerName:     DefaultServerName,
		Logger:         mlog.New(),
		ReadTimeout:    time.Second * 60,
		WriteTimeout:   time.Second * 60,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
}

// SetAddress 设置服务器监听地址
func (s *Server) SetAddress(addr string) {
	s.config.Address = addr
}

// GetAddress 获取服务器监听地址
func (s *Server) GetAddress() string {
	return s.config.Address
}

// SetServeName 设置服务名称
func (s *Server) SetServerName(name string) {
	s.config.ServerName = name
}

// GetServeName 获取服务名称
func (s *Server) GetServerName() string {
	return s.config.ServerName
}

// SetServerRoot 设置服务根目录
func (s *Server) SetServerRoot(root string) {
	s.config.ServerRoot = root
}

// GetServerRoot 获取服务根目录
func (s *Server) GetServerRoot() string {
	return s.config.ServerRoot
}

// Logger 获取日志实例
func (s *Server) Logger() *mlog.Logger {
	return s.config.Logger
}

// SetTimeouts 设置超时配置
func (s *Server) SetTimeouts(read, write, idle time.Duration) {
	s.config.ReadTimeout = read
	s.config.WriteTimeout = write
	s.config.IdleTimeout = idle
}

// GetReadTimeout 获取读取超时时间
func (s *Server) GetReadTimeout() time.Duration {
	return s.config.ReadTimeout
}

// GetWriteTimeout 获取写入超时时间
func (s *Server) GetWriteTimeout() time.Duration {
	return s.config.WriteTimeout
}

// GetIdleTimeout 获取空闲超时时间
func (s *Server) GetIdleTimeout() time.Duration {
	return s.config.IdleTimeout
}

// SetMaxHeaderBytes 设置最大请求头大小
func (s *Server) SetMaxHeaderBytes(size int) {
	s.config.MaxHeaderBytes = size
}

// GetMaxHeaderBytes 获取最大请求头大小
func (s *Server) GetMaxHeaderBytes() int {
	return s.config.MaxHeaderBytes
}
