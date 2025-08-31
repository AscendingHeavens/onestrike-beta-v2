package middleware

import (
	"log"
	"time"

	"github.com/Rishi-Mishra0704/OneStrike/server"
)

// Logging
func Logger() Middleware {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			start := time.Now()
			resp := next(c)
			duration := time.Since(start)
			log.Printf("[%s] %s %s %d (%v)", time.Now().Format(time.RFC3339), c.Request.Method, c.Request.URL.Path, resp.Code, duration)
			return resp
		}
	}
}

// Panic Recovery
func Recovery() Middleware {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered panic: %v", r)
				}
			}()
			return next(c)
		}
	}
}

// Profiling
func ProfilingMiddleware() Middleware {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			start := time.Now()
			resp := next(c)
			log.Printf("Route %s took %v", c.Request.URL.Path, time.Since(start))
			return resp
		}
	}
}
