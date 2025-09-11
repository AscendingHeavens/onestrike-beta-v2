package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AscendingHeavens/onestrike/v2/server"
	"github.com/stretchr/testify/assert"
)

func newCorsTestContext(method, origin string) *server.Context {
	req := httptest.NewRequest(method, "/", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	return &server.Context{
		Request: req,
		Writer:  w,
	}
}

func TestCORS_DefaultConfig_AllowsOrigin(t *testing.T) {
	called := false
	c := newCorsTestContext(http.MethodGet, "http://example.com")
	m := CORS()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.True(t, called, "next handler should be called for non-OPTIONS request")
	assert.Equal(t, http.StatusOK, resp.Code)

	headers := c.Writer.Header()
	assert.Equal(t, "http://example.com", headers.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, headers.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, headers.Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
}

func TestCORS_PreflightRequest_SetsNoContentAndStopsChain(t *testing.T) {
	called := false
	c := newCorsTestContext(http.MethodOptions, "http://example.com")
	m := CORS()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.False(t, called, "next handler should NOT be called for preflight requests")
	assert.Equal(t, http.StatusNoContent, resp.Code)

	// FIX: type assert to *httptest.ResponseRecorder
	recorder, ok := c.Writer.(*httptest.ResponseRecorder)
	assert.True(t, ok, "c.Writer must be a *httptest.ResponseRecorder in tests")
	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestCORS_CustomConfig_SpecificOriginAllowed(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"https://allowed.com"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
		AllowHeaders: []string{"X-Custom-Header"},
	}
	m := CORSWithConfig(cfg)

	called := false
	c := newCorsTestContext(http.MethodGet, "https://allowed.com")

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	handler(c)

	headers := c.Writer.Header()
	assert.Equal(t, "https://allowed.com", headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, strings.Join(cfg.AllowMethods, ", "), headers.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, strings.Join(cfg.AllowHeaders, ", "), headers.Get("Access-Control-Allow-Headers"))
	assert.True(t, called)
}

func TestCORS_CustomConfig_OriginNotAllowed(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"https://allowed.com"},
	}
	m := CORSWithConfig(cfg)

	c := newCorsTestContext(http.MethodGet, "https://disallowed.com")

	handler := m(func(ctx *server.Context) *server.Response {
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	handler(c)

	headers := c.Writer.Header()
	// Should not echo back disallowed origin
	assert.Empty(t, headers.Get("Access-Control-Allow-Origin"))
}
