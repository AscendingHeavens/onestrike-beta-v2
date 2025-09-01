package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Rishi-Mishra0704/OneStrike/server"
)

var (
	ErrCSRFInvalid = errors.New("invalid CSRF token")
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

// Set default CORS config
var defaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodPatch,
		http.MethodPost,
		http.MethodDelete,
	},
	AllowHeaders: []string{
		"Content-Type",
		"Authorization",
		"Accept",
		"Origin",
		"X-Requested-With",
	},
}

// CORS returns a middleware that sets CORS headers.
func CORS() Middleware {
	return CORSWithConfig(defaultCORSConfig)
}

// CORSWithConfig returns a CORS middleware with custom configuration.
func CORSWithConfig(cfg CORSConfig) Middleware {
	// Defaults
	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = []string{"*"}
	}
	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"Content-Type", "Authorization"}
	}

	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			origin := c.Request.Header.Get("Origin")

			// Match origin
			if origin != "" {
				for _, o := range cfg.AllowOrigins {
					if o == "*" || strings.EqualFold(o, origin) {
						c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			// Preflight request
			if c.Request.Method == http.MethodOptions {
				c.Writer.WriteHeader(http.StatusNoContent)
				c.Handled = true
				return &server.Response{Success: true, Message: "CORS preflight", Code: http.StatusNoContent}
			}

			// Continue normal flow
			return next(c)
		}
	}
}

var DefaultCSRFConfig = CSRFConfig{
	TokenHeader:    "X-CSRF-Token",
	TokenCookie:    "csrf_token",
	ContextKey:     "csrf_token",
	Expiry:         24 * time.Hour,
	Secret:         []byte("supersecretkey"),
	SkipMethods:    []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
	CookieSecure:   true,
	CookieHTTPOnly: true,
}

// CSRF returns middleware with default config
func CSRF() Middleware {
	return CSRFWithConfig(DefaultCSRFConfig)
}

// CSRFWithConfig returns middleware with custom config
func CSRFWithConfig(cfg CSRFConfig) Middleware {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			// Skip safe methods
			for _, m := range cfg.SkipMethods {
				if c.Request.Method == m {
					return next(c)
				}
			}

			// Read token from header or cookie
			clientToken := c.Request.Header.Get(cfg.TokenHeader)
			if clientToken == "" {
				if cookie, err := c.Request.Cookie(cfg.TokenCookie); err == nil {
					clientToken = cookie.Value
				}
			}

			// Validate token
			if clientToken != "" {
				serverToken := getOrCreateCSRFToken(c, cfg)
				if !validateCSRFToken(cfg.Secret, serverToken, clientToken) {
					if cfg.ErrorHandler != nil {
						return cfg.ErrorHandler(c, ErrCSRFInvalid)
					}
					return c.String(http.StatusForbidden, ErrCSRFInvalid.Error())
				}
			}

			// Ensure token exists in context and cookie
			token := getOrCreateCSRFToken(c, cfg)
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     cfg.TokenCookie,
				Value:    token,
				Path:     "/",
				Expires:  time.Now().Add(cfg.Expiry),
				Secure:   cfg.CookieSecure,
				HttpOnly: cfg.CookieHTTPOnly,
			})
			c.Params[cfg.ContextKey] = token

			return next(c)
		}
	}
}
