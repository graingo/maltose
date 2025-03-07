package mhttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 静态文件服务增强
func (s *Server) SetStaticPath(prefix string, directory string) {
	// 实现静态文件服务
	s.engine.StaticFS(prefix, http.Dir(directory))
}

// Run 启动 HTTP 服务
func (s *Server) Run() {
	ctx := context.Background()

	// 注册 OpenAPI 和 Swagger
	s.registerDoc(ctx)

	// 在启动前注册所有路由
	s.bindRoutes(ctx)

	// 打印路由信息
	s.printRoute(ctx)

	srv := &http.Server{
		Addr:           s.config.Address,
		Handler:        s.engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	// 创建错误通道
	errChan := make(chan error, 1)
	go func() {
		var err error
		if s.config.TLSEnable {
			if s.config.TLSCertFile == "" || s.config.TLSKeyFile == "" {
				errChan <- fmt.Errorf("TLS certificate and key files are required")
				return
			}
			err = srv.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		s.Logger().Errorf(ctx, "HTTP server %s start failed: %v", s.config.ServerName, err)
	case <-quit:
		s.Logger().Infof(ctx, "Shutting down server...")

		timeout := 5 * time.Second
		if s.config.GracefulEnable {
			timeout = s.config.GracefulTimeout
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if s.config.GracefulEnable {
			// 等待活跃连接完成
			time.Sleep(s.config.GracefulWaitTime)
		}

		if err := srv.Shutdown(ctx); err != nil {
			s.Logger().Errorf(ctx, "Server forced to shutdown: %v", err)
		}
	}
}
