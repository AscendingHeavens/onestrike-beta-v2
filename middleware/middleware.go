package middleware

import (
	"log"
	"runtime"
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

// ProfilingMiddleware logs detailed timing info including handler execution and memory usage
func ProfilingMiddleware() Middleware {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			start := time.Now()

			// Run handler
			resp := next(c)

			// Calculate elapsed
			elapsed := time.Since(start)

			// Capture some runtime stats (GC, mem)
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			log.Printf(
				"[PROFILE] Route: %s | Status: %d | Time: %v | Alloc: %dKB | Sys: %dKB | NumGC: %d",
				c.Request.URL.Path,
				resp.Code,
				elapsed,
				memStats.Alloc/1024,
				memStats.Sys/1024,
				memStats.NumGC,
			)

			return resp
		}
	}
}
