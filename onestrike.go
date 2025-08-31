package onestrike

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Rishi-Mishra0704/OneStrike/middleware"
	"github.com/Rishi-Mishra0704/OneStrike/server"
)

// New creates a new OneStrike Server instance with an empty router and middleware stack.
func New() *Server {
	return &Server{
		router:      server.NewRouter(),
		middlewares: make([]middleware.Middleware, 0),
	}
}

// Use registers a global middleware that will run on every request.
func (s *Server) Use(mw middleware.Middleware) {
	s.middlewares = append(s.middlewares, mw)
}

// UseIf registers a conditional middleware that only runs if the request path
// matches the given pattern. Patterns can include a wildcard '*' at the end.
// Example: UseIf("/api/v1/*", AuthMiddleware())
func (s *Server) UseIf(pattern string, mw middleware.Middleware) {
	s.conditionalMiddleware = append(s.conditionalMiddleware, middleware.ConditionalMiddleware{
		Pattern:    pattern,
		Middleware: mw,
	})
}

// Handle registers a route with a specific HTTP method and path.
// Global middleware is automatically applied in reverse order (so execution order is correct).
func (s *Server) Handle(method, path string, handler server.HandlerFunc) {
	combined := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		combined = s.middlewares[i](combined)
	}
	s.router.Handle(method, path, combined)
}

// Convenience methods for each HTTP method.
func (s *Server) GET(path string, handler server.HandlerFunc) {
	s.Handle(http.MethodGet, path, handler)
}
func (s *Server) POST(path string, handler server.HandlerFunc) {
	s.Handle(http.MethodPost, path, handler)
}
func (s *Server) PUT(path string, handler server.HandlerFunc) {
	s.Handle(http.MethodPut, path, handler)
}
func (s *Server) PATCH(path string, handler server.HandlerFunc) {
	s.Handle(http.MethodPatch, path, handler)
}
func (s *Server) DELETE(path string, handler server.HandlerFunc) {
	s.Handle(http.MethodDelete, path, handler)
}

// ServeHTTP implements http.Handler, so OneStrike Server can be passed
// directly to http.ListenAndServe. It finds the route, applies conditional middleware,
// executes the handler, and writes the Response as JSON.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &server.Context{Writer: w, Request: r}

	// Find the matching handler and path parameters
	handler, params := s.router.FindHandler(r.Method, r.URL.Path)
	if handler == nil {
		http.NotFound(w, r)
		return
	}
	c.Params = params

	// Apply conditional middleware if the request path matches any pattern
	final := handler
	for _, cm := range s.conditionalMiddleware {
		if strings.HasPrefix(r.URL.Path, strings.TrimSuffix(cm.Pattern, "*")) {
			final = cm.Middleware(final)
		}
	}

	// Execute the handler
	resp := final(c)

	// Write JSON response
	if resp != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.Code)
		json.NewEncoder(w).Encode(resp)
	}
}

// Start runs the HTTP server on the specified address. It logs the startup
// and will terminate the program if ListenAndServe returns an error.
func (s *Server) Start(addr string) {
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, s); err != nil {
		log.Fatal(err)
	}
}
