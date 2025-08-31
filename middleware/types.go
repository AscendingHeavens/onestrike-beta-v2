package middleware

import "github.com/Rishi-Mishra0704/OneStrike/server"

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
