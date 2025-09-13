package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// String writes plain text and returns a Response
// Example: c.String(200, "Hello World")
func (c *Context) String(code int, s string) *Response {
	c.writeResponse(code, "text/plain", []byte(s))
	return &Response{Success: true, Message: s, Code: code}
}

// HTML writes HTML content and returns a Response
// Example: c.HTML(200, "<h1>Hello</h1>")
func (c *Context) HTML(code int, html string) *Response {
	c.writeResponse(code, "text/html", []byte(html))
	return &Response{Success: true, Message: "HTML written", Code: code}
}

// Blob writes raw binary data with the given Content-Type and returns a Response
// Example: c.Blob(200, data, "image/png")
func (c *Context) Blob(code int, data []byte, contentType string) *Response {
	c.writeResponse(code, contentType, data)
	return &Response{Success: true, Message: "Blob written", Code: code}
}

// JSON writes the given Response object as JSON with the provided status code.
// This method respects c.Handled, so it won't write twice if something else already wrote.
func (c *Context) JSON(success bool, message string, details any, code int) *Response {
	resp := &Response{
		Success: success,
		Message: message,
		Details: details,
		Code:    code,
	}
	if c.Handled {
		return resp
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	_ = json.NewEncoder(c.Writer).Encode(resp)
	c.Handled = true
	return resp
}

// JSON writes the given Response object as JSON with the provided status code.
// This method respects c.Handled, so it won't write twice if something else already wrote.
func (c *Context) ErrorJSON(message string, details any, code int) *Response {
	resp := &Response{
		Success: false,
		Message: message,
		Details: details,
		Code:    code,
	}
	if c.Handled {
		return resp
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	_ = json.NewEncoder(c.Writer).Encode(resp)
	c.Handled = true
	return resp
}

// Redirect sends an HTTP redirect to the specified location.
func (c *Context) Redirect(code int, location string) *Response {
	if c.Handled {
		return &Response{Success: false, Message: "Response already handled", Code: code}
	}
	c.Writer.Header().Set("Location", location)
	c.Writer.WriteHeader(code)
	c.Handled = true
	return &Response{
		Success: true,
		Message: "Redirected to " + location,
		Code:    code,
	}
}

// File serves a file from disk with proper Content-Type.
// If file doesn't exist or can't be read, returns a 404/500 JSON response.
func (c *Context) File(filePath string) *Response {
	if c.Handled {
		return &Response{Success: false, Message: "Response already handled", Code: 500}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return &Response{Success: false, Message: "File not found", Code: 404}
	}

	contentType := http.DetectContentType(data)
	c.writeResponse(200, contentType, data)

	return &Response{
		Success: true,
		Message: fmt.Sprintf("Served file: %s", filePath),
		Code:    200,
	}
}
