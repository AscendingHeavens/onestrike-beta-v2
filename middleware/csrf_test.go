package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rishi-Mishra0704/OneStrike/server"
	"github.com/stretchr/testify/assert"
)

func newCSRFTContext(method string, token string) *server.Context {
	req := httptest.NewRequest(method, "/", nil)
	if token != "" {
		req.Header.Set(DefaultCSRFConfig.TokenHeader, token)
	}
	w := httptest.NewRecorder()
	return &server.Context{
		Request: req,
		Writer:  w,
		Params:  make(map[string]string),
	}
}

func TestCSRF_SkipSafeMethods(t *testing.T) {
	called := false
	c := newCSRFTContext(http.MethodGet, "")
	m := CSRF()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.True(t, called, "handler must be called for safe method")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestCSRF_SetsTokenCookie_WhenMissing(t *testing.T) {
	called := false
	c := newCSRFTContext(http.MethodPost, "")
	m := CSRF()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.True(t, called, "handler must be called after token creation")
	assert.Equal(t, http.StatusOK, resp.Code)

	// Verify token is set in context
	token := c.Params[DefaultCSRFConfig.ContextKey]
	assert.NotEmpty(t, token, "CSRF token should be generated and stored in context")

	// Verify cookie is set
	recorder := c.Writer.(*httptest.ResponseRecorder)
	cookies := recorder.Result().Cookies()
	assert.NotEmpty(t, cookies, "CSRF cookie must be set")
	found := false
	for _, ck := range cookies {
		if ck.Name == DefaultCSRFConfig.TokenCookie {
			found = true
			assert.Equal(t, token, ck.Value)
			assert.True(t, ck.HttpOnly)
			assert.True(t, ck.Secure)
		}
	}
	assert.True(t, found, "CSRF token cookie must be present")
}

func TestCSRF_SetsTokenCookie_OnPostWithoutToken(t *testing.T) {
	called := false
	c := newCSRFTContext(http.MethodPost, "")
	m := CSRF()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.True(t, called, "handler must be called")
	assert.Equal(t, http.StatusOK, resp.Code)

	// Cookie should be set
	rec := c.Writer.(*httptest.ResponseRecorder)
	cookies := rec.Result().Cookies()
	assert.NotEmpty(t, cookies, "csrf cookie must be set")
}

func TestCSRF_InvalidToken_Returns403(t *testing.T) {
	called := false
	c := newCSRFTContext(http.MethodPost, "fake-token")
	m := CSRF()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.False(t, called, "handler must NOT be called for invalid token")
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestCSRF_InvalidToken_BlocksRequest(t *testing.T) {
	called := false
	c := newCSRFTContext(http.MethodPost, "tampered-token")
	m := CSRF()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.False(t, called, "handler must NOT be called for invalid token")
	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Contains(t, resp.Message, "invalid CSRF token")
}

func testGenerateCSRFToken(c *server.Context, cfg CSRFConfig) string {
	return getOrCreateCSRFToken(c, cfg)
}
