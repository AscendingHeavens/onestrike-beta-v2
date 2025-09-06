package server

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const (
	maxBodySize   = 10 << 20 // 10MB
	maxMemorySize = 32 << 20 // 32MB for multipart
)

// Bind reads request body and decodes based on Content-Type.
// Automatically writes 400 response on error.
func (c *Context) Bind(dest any) error {
	if err := c.ShouldBind(dest); err != nil {
		c.writeErrorResponse(http.StatusBadRequest, "Invalid request body", err)
		return err
	}
	return nil
}

// BindJSON enforces JSON Content-Type and decodes into dest.
// Automatically writes 400 response on error.
func (c *Context) BindJSON(dest any) error {
	if err := c.ShouldBindJSON(dest); err != nil {
		c.writeErrorResponse(http.StatusBadRequest, "Invalid JSON body", err)
		return err
	}
	return nil
}

// BindXML enforces XML Content-Type and decodes into dest.
// Automatically writes 400 response on error.
func (c *Context) BindXML(dest any) error {
	if err := c.ShouldBindXML(dest); err != nil {
		c.writeErrorResponse(http.StatusBadRequest, "Invalid XML body", err)
		return err
	}
	return nil
}

// ShouldBind attempts to bind based on Content-Type without writing response.
func (c *Context) ShouldBind(dest any) error {
	contentType := c.Request.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("missing Content-Type header")
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("invalid Content-Type: %w", err)
	}

	switch mediaType {
	case "application/json":
		return c.ShouldBindJSON(dest)
	case "application/xml", "text/xml":
		return c.ShouldBindXML(dest)
	case "application/x-www-form-urlencoded":
		return c.ShouldBindForm(dest)
	case "multipart/form-data":
		return c.ShouldBindMultipart(dest)
	default:
		return fmt.Errorf("unsupported Content-Type: %s", mediaType)
	}
}

// ShouldBindJSON decodes JSON without automatic error response.
func (c *Context) ShouldBindJSON(dest any) error {
	return c.shouldBindBody(dest, "application/json", json.Unmarshal)
}

// ShouldBindXML decodes XML without automatic error response.
func (c *Context) ShouldBindXML(dest any) error {
	return c.shouldBindBody(dest, "application/xml", xml.Unmarshal)
}

func (c *Context) ShouldBindForm(dest any) error {
	if err := c.Request.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}
	if len(c.Request.PostForm) == 0 {
		return errors.New("no form values found")
	}
	return c.bindFormToStruct(c.Request.PostForm, dest)
}

// ShouldBindMultipart binds multipart form to struct.
func (c *Context) ShouldBindMultipart(dest any) error {
	if err := c.Request.ParseMultipartForm(maxMemorySize); err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}
	return c.bindFormToStruct(c.Request.MultipartForm.Value, dest)
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

// FormFile retrieves the file from a multipart form.
// It returns the first file for the provided form key.
// Example: <input type="file" name="avatar" /> -> c.FormFile("avatar")
func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(name)
}

// FormValue returns the first value for the named component of the POST or PUT request body.
// It calls ParseMultipartForm and ParseForm if necessary.
// Example: <input type="text" name="username" /> -> c.FormValue("username")
func (c *Context) FormValue(name string) string {
	return c.Request.FormValue(name)
}

// BindForm parses form data and binds it to a map
// Takes only the first value for each key
func (c *Context) BindForm(dest map[string]string) error {
	// Check if content type is form URL-encoded (allowing charset parameter)
	contentType := c.Request.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("missing content type header")
	}

	// Parse media type to handle charset parameter
	mediaType := strings.ToLower(strings.Split(contentType, ";")[0])
	if strings.TrimSpace(mediaType) != "application/x-www-form-urlencoded" {
		return fmt.Errorf("invalid content type, expected application/x-www-form-urlencoded, got %s", mediaType)
	}

	// Parse the form (this handles both POST body and URL query parameters)
	if err := c.Request.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	// Copy form values into dest map (first value only)
	for key, values := range c.Request.PostForm {
		if len(values) > 0 {
			dest[key] = values[0]
		}
	}

	return nil
}

// BindFormAll parses form data and binds all values to a url.Values map
func (c *Context) BindFormAll() (url.Values, error) {
	// Check content type
	contentType := c.Request.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("missing content type header")
	}

	mediaType := strings.ToLower(strings.Split(contentType, ";")[0])
	if strings.TrimSpace(mediaType) != "application/x-www-form-urlencoded" {
		return nil, fmt.Errorf("invalid content type, expected application/x-www-form-urlencoded, got %s", mediaType)
	}

	// Parse the form
	if err := c.Request.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	// Return all form values
	return c.Request.PostForm, nil
}
