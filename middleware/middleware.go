package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
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
// Recovery returns a middleware that recovers from panics and writes a 500 response.
func Recovery() Middleware {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			defer func() {
				if r := recover(); r != nil {
					// Log stack trace
					log.Printf("Recovered panic: %v\n%s", r, string(debug.Stack()))

					// Only write response if handler hasn't already written
					if !c.Handled {
						accept := c.Request.Header.Get("Accept")
						if strings.Contains(accept, "text/html") {
							c.HTML(http.StatusInternalServerError, "<h1>500 Internal Server Error</h1>")
						} else {
							c.JSON(http.StatusInternalServerError, &server.Response{
								Success: false,
								Message: "Internal Server Error",
								Code:    http.StatusInternalServerError,
								Details: fmt.Sprintf("%v", r),
							})
						}
					}
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
