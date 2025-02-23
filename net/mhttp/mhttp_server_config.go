package mhttp

import "github.com/mingzaily/maltose/os/mlog"

// ServerConfig 服务器配置
type ServerConfig struct {
	Address    string
	ServerName string
	ServerRoot string
	Logger     *mlog.Logger
}

func NewConfig() ServerConfig {
	return ServerConfig{
		Address:    defaultPort,
		ServerName: DefaultServerName,
		Logger:     mlog.New(),
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
