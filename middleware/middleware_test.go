package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ascending-Heavens/onestrike/server"
	"github.com/stretchr/testify/assert"
)

func newTestContext(method string) *server.Context {
	req := httptest.NewRequest(method, "/", nil)
	w := httptest.NewRecorder()
	return &server.Context{
		Request: req,
		Writer:  w,
		Params:  make(map[string]string),
	}
}

func TestLogger_CallsNextAndReturnsResponse(t *testing.T) {
	called := false
	c := newTestContext(http.MethodGet)
	m := Logger()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.True(t, called, "logger must call next handler")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRecovery_NoPanic_PassesThrough(t *testing.T) {
	called := false
	c := newTestContext(http.MethodGet)
	m := Recovery()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusOK}
	})

	resp := handler(c)

	assert.True(t, called, "recovery must call next handler when no panic")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRecovery_Panic_Returns500_JSON(t *testing.T) {
	c := newTestContext(http.MethodGet)
	m := Recovery()

	handler := m(func(ctx *server.Context) *server.Response {
		panic("boom")
	})

	// This will trigger recovery middleware
	resp := handler(c)

	// resp will be nil because middleware wrote response manually
	assert.Nil(t, resp)

	// Now check the recorder output
	rec := c.Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Internal Server Error")
	assert.Contains(t, body, "boom")
}

func TestRecovery_Panic_ReturnsHTML_WhenAcceptHeaderSet(t *testing.T) {
	c := newTestContext(http.MethodGet)
	c.Request.Header.Set("Accept", "text/html")
	m := Recovery()

	handler := m(func(ctx *server.Context) *server.Response {
		panic("boom")
	})

	resp := handler(c)

	assert.Nil(t, resp)

	rec := c.Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "<h1>500 Internal Server Error</h1>")
}

func TestProfilingMiddleware_CallsNextAndReturnsResponse(t *testing.T) {
	called := false
	c := newTestContext(http.MethodGet)
	m := ProfilingMiddleware()

	handler := m(func(ctx *server.Context) *server.Response {
		called = true
		return &server.Response{Success: true, Code: http.StatusCreated}
	})

	resp := handler(c)

	assert.True(t, called, "profiling middleware must call next handler")
	assert.Equal(t, http.StatusCreated, resp.Code)
}

func TestRecovery_Panic_WhenAlreadyHandled_DoesNotWriteAgain(t *testing.T) {
	c := newTestContext(http.MethodGet)
	c.Handled = true // simulate handler already wrote a response
	m := Recovery()

	handler := m(func(ctx *server.Context) *server.Response {
		panic("already handled panic")
	})

	resp := handler(c)
	assert.Nil(t, resp)

	rec := c.Writer.(*httptest.ResponseRecorder)
	// Should NOT write anything new, response should remain empty
	assert.Equal(t, 200, rec.Code) // Recorder defaults to 200 if not written
	assert.Empty(t, rec.Body.String())
}
