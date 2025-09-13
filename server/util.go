package server

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
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

// writeErrorResponse writes standardized error response.
func (c *Context) writeErrorResponse(code int, message string, err error) {
	if c.Handled {
		return
	}

	c.JSON(false,
		message,
		err.Error(),
		code)
}

// shouldBindBody is the core implementation for body binding with size limits.
func (c *Context) shouldBindBody(dest any, expectedType string, unmarshal func([]byte, any) error) error {
	// Validate Content-Type if specified
	if expectedType != "" {
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "" {
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				return fmt.Errorf("invalid Content-Type: %w", err)
			}
			if mediaType != expectedType && (expectedType != "application/xml" || mediaType != "text/xml") {
				return fmt.Errorf("expected Content-Type %s, got %s", expectedType, mediaType)
			}
		}
	}

	if c.Request.Body == nil {
		return errors.New("request body is empty")
	}

	limitedReader := http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	return unmarshal(body, dest)
}

// bindFormToStruct uses reflection to bind form values to struct fields.
func (c *Context) bindFormToStruct(values url.Values, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("destination must be a pointer to struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get form tag or use field name
		tagName := fieldType.Tag.Get("form")
		if tagName == "" {
			tagName = strings.ToLower(fieldType.Name)
		}
		if tagName == "-" {
			continue
		}

		formValue := values.Get(tagName)
		if formValue == "" {
			continue
		}

		if err := c.setFieldValue(field, formValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue sets a reflect.Value based on its type.
func (c *Context) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
