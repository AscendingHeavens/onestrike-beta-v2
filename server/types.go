package server

import "net/http"

// Response is the unified return type for all handlers in OneStrike.
// It is automatically serialized to JSON and written to the client.
// Fields:
//   - Success: indicates whether the request was successful.
//   - Message: human-readable message describing the result.
//   - Details: optional field containing extra data (any type).
//   - Code: HTTP status code to be sent to the client. It's Required to have at least one Status code
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Code    int    `json:"code"` // required
}

// Context wraps http.ResponseWriter and *http.Request, providing
// convenience access to route parameters and helper methods in the future.
// Fields:
//   - Writer: the http.ResponseWriter to write responses.
//   - Request: the incoming HTTP request.
//   - Params: a map of path parameters extracted from the route (e.g., ":id").
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	Handled bool
}

// HandlerFunc defines the signature for all route handlers in OneStrike.
// Every handler receives a pointer to a Context and returns a pointer to a Response.
// The framework automatically writes the Response as JSON to the client.
type HandlerFunc func(c *Context) *Response
