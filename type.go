package onestrike

import (
	"net/http"

	"github.com/Rishi-Mishra0704/OneStrike/middleware"
	"github.com/Rishi-Mishra0704/OneStrike/server"
)

// Server is the main entry point for the OneStrike framework.
// It holds the router, global middlewares, and any conditional middlewares
// that should be applied based on route patterns.
type Server struct {
	// router is the core request router responsible for mapping HTTP methods
	// and paths to handler functions.
	router *server.Router

	// middlewares is a slice of global middleware that runs on every request
	// before the matched route handler.
	middlewares []middleware.Middleware

	// conditionalMiddleware is a slice of middleware that only run when the
	// incoming request path matches the provided pattern.
	// For example, you might apply authentication middleware only for
	// `/api/*` routes.
	conditionalMiddleware []middleware.ConditionalMiddleware
}

// Group represents a collection of routes that share a common path prefix
// and middleware stack. Useful for organizing related endpoints like `/api/v1/*`.
type Group struct {
	// Prefix is the base path for this group (e.g., "/api/v1").
	Prefix string

	// Server is a reference back to the parent server, allowing
	// groups to register routes directly into the main router.
	Server *Server

	// Middlewares is a list of middleware that will be applied to
	// every route registered within this group, in addition to any
	// global or conditional middleware from the Server.
	Middlewares []middleware.Middleware
}

// Context is an alias to server.Context, which wraps the request and response
// writer and provides convenience methods (params, body parsing, etc.).
type Context = server.Context

// Response is an alias to server.Response, the unified return type
// from every handler function. Encoded as JSON and written to the client.
type Response = server.Response

// Middleware is an alias to middleware.Middleware, representing a function
// that wraps and modifies a HandlerFunc, similar to how middleware works
// in frameworks like Express or Fiber.
type Middleware = middleware.Middleware

// ConditionalMiddleware is an alias to middleware.ConditionalMiddleware,
// which pairs a pattern (e.g., "/api/*") with a Middleware function.
type ConditionalMiddleware = middleware.ConditionalMiddleware

// HandlerFunc is an alias to server.HandlerFunc, the function signature
// that route handlers must implement. It takes a *Context and returns a *Response.
type HandlerFunc = server.HandlerFunc

// Option A: Using an interface (Recommended)

// First, define an interface for the server starter
type TLSStarter interface {
	startTLSServer(*http.Server)
}
