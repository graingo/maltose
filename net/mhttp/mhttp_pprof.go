package mhttp

import (
	"net/http/pprof"
	"strings"
)

const (
	defaultPProfPattern = "/debug/pprof"
)

// utilPProf 是 PProf 接口实现
type utilPProf struct{}

// EnablePProf 为服务器启用 PProf 功能
func (s *Server) EnablePProf(pattern ...string) {
	p := defaultPProfPattern
	if len(pattern) > 0 && pattern[0] != "" {
		p = pattern[0]
	}

	up := &utilPProf{}
	uri := strings.TrimRight(p, "/")

	s.Group(uri, func(group *RouterGroup) {
		group.BindHandler("GET", "/:action", up.Index)
		group.BindHandler("GET", "/cmdline", up.Cmdline)
		group.BindHandler("GET", "/profile", up.Profile)
		group.BindHandler("GET", "/symbol", up.Symbol)
		group.BindHandler("GET", "/trace", up.Trace)
	})
}

// Index 显示 PProf 索引页面
func (p *utilPProf) Index(r *Request) {
	action := r.Param("action")
	if action == "" {
		pprof.Index(r.Writer, r.Request)
		return
	}

	pprof.Handler(action).ServeHTTP(r.Writer, r.Request)
}

// Cmdline 响应运行程序的命令行
func (p *utilPProf) Cmdline(r *Request) {
	pprof.Cmdline(r.Writer, r.Request)
}

// Profile 响应 pprof 格式的 CPU 配置文件
func (p *utilPProf) Profile(r *Request) {
	pprof.Profile(r.Writer, r.Request)
}

// Symbol 查找请求中列出的程序计数器
func (p *utilPProf) Symbol(r *Request) {
	pprof.Symbol(r.Writer, r.Request)
}

// Trace 以二进制形式响应执行跟踪
func (p *utilPProf) Trace(r *Request) {
	pprof.Trace(r.Writer, r.Request)
}
