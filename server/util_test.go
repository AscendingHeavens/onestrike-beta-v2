package server

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- setFieldValue table-driven tests ---
func TestSetFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		field     reflect.Value
		value     string
		wantVal   any
		expectErr bool
	}{
		{"string", reflect.ValueOf(new(string)).Elem(), "hello", "hello", false},
		{"int", reflect.ValueOf(new(int64)).Elem(), "42", int64(42), false},
		{"int_fail", reflect.ValueOf(new(int64)).Elem(), "notanint", nil, true},
		{"bool_true", reflect.ValueOf(new(bool)).Elem(), "true", true, false},
		{"bool_false", reflect.ValueOf(new(bool)).Elem(), "false", false, false},
		{"bool_fail", reflect.ValueOf(new(bool)).Elem(), "oops", nil, true},
		{"float64", reflect.ValueOf(new(float64)).Elem(), "3.14", 3.14, false},
		{"float_fail", reflect.ValueOf(new(float64)).Elem(), "nan", math.NaN(), false},
		{"unsupported", reflect.ValueOf(new([]string)).Elem(), "value", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{}
			err := c.setFieldValue(tt.field, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.name == "float_fail" {
					fVal := tt.field.Interface().(float64)
					assert.True(t, math.IsNaN(fVal))
				} else {
					assert.Equal(t, tt.wantVal, tt.field.Interface())
				}
			}
		})
	}
}

func TestBindFormToStruct(t *testing.T) {
	type payload struct {
		Name  string  `form:"name"`
		Age   int     `form:"age"`
		Admin bool    `form:"admin"`
		Rate  float64 `form:"rate"`
		Skip  string  `form:"-"`
	}
	values := url.Values{
		"name":  {"Rishi"},
		"age":   {"21"},
		"admin": {"true"},
		"rate":  {"3.14"},
	}
	c := &Context{}
	var p payload
	err := c.bindFormToStruct(values, &p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
	assert.Equal(t, 21, p.Age)
	assert.Equal(t, true, p.Admin)
	assert.Equal(t, 3.14, p.Rate)
	assert.Equal(t, "", p.Skip)
}

func TestShouldBindBody(t *testing.T) {
	t.Run("empty body", func(t *testing.T) {
		c := &Context{Request: &http.Request{Body: http.NoBody}, Writer: httptest.NewRecorder()}
		var dest any
		err := c.shouldBindBody(&dest, "application/json", json.Unmarshal)
		assert.Error(t, err)
	})

	t.Run("valid JSON", func(t *testing.T) {
		body := `{"name":"Rishi"}`
		c := &Context{Request: httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), Writer: httptest.NewRecorder()}
		c.Request.Header.Set("Content-Type", "application/json")
		var dest map[string]string
		err := c.shouldBindBody(&dest, "application/json", json.Unmarshal)
		assert.NoError(t, err)
		assert.Equal(t, "Rishi", dest["name"])
	})

	t.Run("content-type mismatch", func(t *testing.T) {
		body := `{"name":"Rishi"}`
		c := &Context{Request: httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), Writer: httptest.NewRecorder()}
		c.Request.Header.Set("Content-Type", "application/xml")
		var dest map[string]string
		err := c.shouldBindBody(&dest, "application/json", json.Unmarshal)
		assert.Error(t, err)
	})
}

func TestWriteResponse(t *testing.T) {
	t.Run("writes when not handled", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := &Context{Writer: rec}
		body := []byte("hello")
		c.writeResponse(200, "text/plain", body)
		assert.Equal(t, 200, rec.Code)
		assert.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
		assert.Equal(t, "hello", rec.Body.String())
		assert.True(t, c.Handled)
	})
	t.Run("skips when handled", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := &Context{Writer: rec, Handled: true}
		c.writeResponse(200, "text/plain", []byte("hi"))

		// Code may default to 200, so check body and headers instead
		assert.Empty(t, rec.Body.String())
		assert.False(t, rec.Header().Get("Content-Type") == "text/plain")
	})

}

func TestWriteErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	c := &Context{Writer: rec}
	err := errors.New("oops")
	c.writeErrorResponse(400, "bad request", err)
	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "bad request")
	assert.Contains(t, rec.Body.String(), "oops")
}
