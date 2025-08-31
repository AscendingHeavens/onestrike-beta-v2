package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// route represents a single registered route in the router.
// It contains the HTTP method, the route path pattern, and the handler function.
type route struct {
	Method  string      // HTTP method (GET, POST, PUT, etc.)
	Path    string      // Route pattern, e.g. "/users/:id"
	Handler HandlerFunc // Function to handle requests matching this route
}

// Router is a minimal HTTP router that supports method-based routing
// and simple path parameters (e.g., /users/:id).
type Router struct {
	routes []route // List of all registered routes
}

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

// Bind reads the request body and attempts to decode it into dest.
// If decoding fails, it automatically writes a 400 Bad Request response
// and returns the error. Use this when you want to fail fast.
func (c *Context) Bind(dest any) error {
	return c.bindBody(dest, false)
}

// BindJSON reads the request body as JSON and attempts to decode it into dest.
// If decoding fails, it automatically writes a 400 Bad Request response
// and returns the error. This enforces Content-Type: application/json.
func (c *Context) BindJSON(dest any) error {
	return c.bindBody(dest, true)
}

// ShouldBind reads the request body and attempts to decode it into dest.
// Unlike Bind, it does NOT write a response automatically on error.
// It simply returns the error, giving you full control over the response.
func (c *Context) ShouldBind(dest any) error {
	return c.shouldBindBody(dest, false)
}

// ShouldBindJSON reads the request body as JSON and attempts to decode it into dest.
// Unlike BindJSON, it does NOT write a response automatically on error.
// It simply returns the error, giving you full control over the response.
func (c *Context) ShouldBindJSON(dest any) error {
	return c.shouldBindBody(dest, true)
}

// bindBody is the internal implementation for Bind and BindJSON.
// It attempts to decode the body into dest, and if an error occurs,
// it writes a standardized JSON error response.
func (c *Context) bindBody(dest any, jsonOnly bool) error {
	if err := c.shouldBindBody(dest, jsonOnly); err != nil {
		resp := &Response{
			Success: false,
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		}
		c.JSON(http.StatusBadRequest, resp)
		return err
	}
	return nil
}

// shouldBindBody is the internal implementation for ShouldBind and ShouldBindJSON.
// It just performs the decode and returns the error, without writing a response.
func (c *Context) shouldBindBody(dest any, jsonOnly bool) error {
	if jsonOnly {
		contentType := c.Request.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return errors.New("expected Content-Type: application/json")
		}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	defer c.Request.Body.Close()

	// Currently only supports JSON. Can be extended for forms, XML, etc.
	return json.Unmarshal(body, dest)
}

// JSON writes the given response as JSON with the provided status code.
func (c *Context) JSON(code int, resp *Response) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	_ = json.NewEncoder(c.Writer).Encode(resp)
}
