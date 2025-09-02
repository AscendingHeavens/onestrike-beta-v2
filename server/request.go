package server

import (
	"mime/multipart"
	"net/url"
)

// Bind reads the request body and attempts to decode it into dest.
// If decoding fails, it automatically writes a 400 Bad Request response
// and returns the error. Use this when you want to fail fast.
func (c *Context) Bind(dest any) error {
	return c.bindBody(dest, false)
}

// BindJSON reads the request body as JSON and attempts to decode it into dest.
// If decoding fails, it automatically writes a 400 Bad Request response
// and returns the error. This enforces Content-Type: application/json.
//
// You typically do NOT need to check the returned error yourself — just return nil
// from your handler to stop further execution. If you want custom error handling,
// use ShouldBindJSON instead.
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

// Param returns the value of a path parameter by name.
// Example: /users/:id -> c.Param("id") returns "123"
func (c *Context) Param(name string) string {
	if c.Params == nil {
		return ""
	}
	return c.Params[name]
}

// Query returns the first value of a URL query parameter by key.
// Example: /search?q=golang -> c.Query("q") returns "golang"
func (c *Context) Query(key string) string {
	if c.Request == nil {
		return ""
	}
	return c.Request.URL.Query().Get(key)
}

// QueryArray returns all values for a query parameter key.
// Example: /filter?tag=go&tag=web -> c.QueryArray("tag") returns []string{"go", "web"}
func (c *Context) QueryArray(key string) []string {
	if c.Request == nil {
		return nil
	}
	values, _ := url.ParseQuery(c.Request.URL.RawQuery)
	return values[key]
}

func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(name)
}

func (c *Context) FormValue(name string) string {
	return c.Request.FormValue(name)
}
