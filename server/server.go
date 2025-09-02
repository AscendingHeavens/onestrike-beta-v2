package server

import (
	"strings"
)

// NewRouter creates and returns a new Router instance.
func NewRouter() *Router {
	return &Router{routes: make([]route, 0)}
}

// Handle registers a new route with a specific HTTP method, path, and handler.
func (r *Router) Handle(method, path string, handler HandlerFunc) {
	r.routes = append(r.routes, route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

// FindHandler attempts to match an incoming request (method + path)
// against the registered routes. It supports simple path parameters
// like "/users/:id" and extracts them into a map.
// Returns the matching HandlerFunc and a map of extracted params.
// If no match is found, it returns (nil, nil).
func (r *Router) FindHandler(method, path string) (HandlerFunc, map[string]string) {
	for _, rt := range r.routes {
		// Skip if method doesn't match
		if rt.Method != method {
			continue
		}

		// Split both route and incoming path into parts
		params := make(map[string]string)
		rtParts := strings.Split(rt.Path, "/")
		pParts := strings.Split(path, "/")

		// Length mismatch -> no match
		if len(rtParts) != len(pParts) {
			continue
		}

		// Check each segment
		match := true
		for i := range rtParts {
			if strings.HasPrefix(rtParts[i], ":") {
				// It's a path parameter, capture it
				params[rtParts[i][1:]] = pParts[i]
			} else if rtParts[i] != pParts[i] {
				// Static segment mismatch -> route doesn't match
				match = false
				break
			}
		}

		if match {
			return rt.Handler, params
		}
	}

	// No matching route found
	return nil, nil
}
