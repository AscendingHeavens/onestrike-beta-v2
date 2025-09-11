package middleware

import (
	"time"

	"github.com/AscendingHeavens/onestrike/server"
)

// Middleware defines the function signature for all middleware in OneStrike.
// A middleware wraps a HandlerFunc, allowing pre- or post-processing of requests.
// Example: logging, authentication, profiling, or panic recovery.
type Middleware func(server.HandlerFunc) server.HandlerFunc

// ConditionalMiddleware pairs a middleware with a path pattern.
// The middleware is only applied if the request path matches the pattern.
// Patterns can use a wildcard '*' at the end to match any subpath.
type ConditionalMiddleware struct {
	Pattern    string     // The URL path pattern to match, e.g., "/api/v1/*"
	Middleware Middleware // The middleware function to apply when the pattern matches
}

// CORSConfig defines allowed origins, headers, and methods.
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

type CSRFConfig struct {
	TokenHeader    string        // header to read/write token
	TokenCookie    string        // cookie name
	ContextKey     string        // context key for token
	Expiry         time.Duration // token expiry
	Secret         []byte        // HMAC secret
	SkipMethods    []string      // methods that don't require validation
	ErrorHandler   func(*server.Context, error) *server.Response
	CookieSecure   bool
	CookieHTTPOnly bool
}
