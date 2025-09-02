package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (c *Context) writeResponse(code int, contentType string, body []byte) {
	if c.Handled {
		return
	}
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.WriteHeader(code)
	_, _ = c.Writer.Write(body)
	c.Handled = true
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
