package onestrike

import (
	"net/http"

	"github.com/Ascending-Heavens/onestrike/middleware"
)

// Group represents a collection of routes sharing a common prefix
// and middleware stack. Useful for organizing related endpoints.
// Example: v1 := app.Group("/api/v1")
func (s *Server) Group(prefix string) *Group {
	return &Group{
		Prefix:      prefix,
		Server:      s,
		Middlewares: make([]middleware.Middleware, 0),
	}
}

// Use registers a middleware for this specific group.
// These middlewares are applied only to routes within the group,
// in addition to any global middleware from the parent server.
func (g *Group) Use(mw middleware.Middleware) {
	g.Middlewares = append(g.Middlewares, mw)
}

// Handle registers a route for the group with a specific HTTP method and path.
// It automatically prepends the group's prefix to the path and applies
// the group's middleware stack in reverse order for correct execution.
func (g *Group) Handle(method, path string, handler HandlerFunc) {
	fullPath := g.Prefix + path

	combined := handler

	// Apply server-level middlewares first
	for i := len(g.Server.middlewares) - 1; i >= 0; i-- {
		combined = g.Server.middlewares[i](combined)
	}

	// Then group-specific middlewares
	for i := len(g.Middlewares) - 1; i >= 0; i-- {
		combined = g.Middlewares[i](combined)
	}

	g.Server.router.Handle(method, fullPath, combined)
}

// Convenience methods for common HTTP methods for group routes.
func (g *Group) GET(path string, handler HandlerFunc)    { g.Handle(http.MethodGet, path, handler) }
func (g *Group) POST(path string, handler HandlerFunc)   { g.Handle(http.MethodPost, path, handler) }
func (g *Group) PUT(path string, handler HandlerFunc)    { g.Handle(http.MethodPut, path, handler) }
func (g *Group) PATCH(path string, handler HandlerFunc)  { g.Handle(http.MethodPatch, path, handler) }
func (g *Group) DELETE(path string, handler HandlerFunc) { g.Handle(http.MethodDelete, path, handler) }
