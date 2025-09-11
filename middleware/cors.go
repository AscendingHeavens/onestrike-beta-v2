package middleware

import (
	"net/http"
	"strings"

	"github.com/AscendingHeavens/onestrike/v2/server"
)

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
