package mhttp

import (
	"net/http/pprof"
	"strings"
)

const (
	defaultPProfPattern = "/debug/pprof"
)

// utilPProf is the PProf interface implementation
type utilPProf struct{}

// EnablePProf enables PProf functionality for the server
func (s *Server) EnablePProf(pattern ...string) {
	p := defaultPProfPattern
	if len(pattern) > 0 && pattern[0] != "" {
		p = pattern[0]
	}

	up := &utilPProf{}
	uri := strings.TrimRight(p, "/")

	s.Group(uri, func(group *RouterGroup) {
		group.GET("/:action", up.Index)
		group.GET("/cmdline", up.Cmdline)
		group.GET("/profile", up.Profile)
		group.GET("/symbol", up.Symbol)
		group.GET("/trace", up.Trace)
	})
}

// Index displays the PProf index page
func (p *utilPProf) Index(r *Request) {
	action := r.Param("action")
	if action == "" {
		pprof.Index(r.Writer, r.Request)
		return
	}

	pprof.Handler(action).ServeHTTP(r.Writer, r.Request)
}

// Cmdline responds to the command line of the running program
func (p *utilPProf) Cmdline(r *Request) {
	pprof.Cmdline(r.Writer, r.Request)
}

// Profile responds to the CPU profile in pprof format
func (p *utilPProf) Profile(r *Request) {
	pprof.Profile(r.Writer, r.Request)
}

// Symbol finds the program counter in the request
func (p *utilPProf) Symbol(r *Request) {
	pprof.Symbol(r.Writer, r.Request)
}

// Trace responds to the execution trace in binary format
func (p *utilPProf) Trace(r *Request) {
	pprof.Trace(r.Writer, r.Request)
}
