package mhttp

import (
	"time"
)

// MiddlewareLog is a middleware for logging HTTP requests.
func MiddlewareLog() MiddlewareFunc {
	return func(r *Request) {
		// start time
		start := time.Now()

		// get request information
		path := r.Request.URL.Path
		raw := r.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// execute next middleware
		r.Next()

		// calculate processing time
		latency := time.Since(start)

		// get response status
		status := r.Writer.Status()

		// record log
		r.Logger().Infof(r.Request.Context(),
			"[HTTP] %-3d | %13v | %-15s | %-7s | %s",
			status,           // status code fixed 3 digits
			latency,          // latency fixed 13 digits
			r.ClientIP(),     // IP address fixed 15 digits
			r.Request.Method, // HTTP method fixed 7 digits
			path,
		)

		// if there are errors, record error log
		if len(r.Errors) > 0 {
			for _, e := range r.Errors {
				r.Logger().Errorf(r.Request.Context(),
					"[HTTP] %s | Error: %v",
					r.GetServerName(),
					e.Err,
				)
			}
		}
	}
}
