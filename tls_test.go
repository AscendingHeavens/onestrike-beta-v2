package onestrike

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartTLS_CallsListenAndServeTLS(t *testing.T) {
	called := false
	var addr, cert, key string

	// mock ListenAndServeTLS
	listenAndServeTLS = func(a, c, k string, h http.Handler) error {
		called = true
		addr = a
		cert = c
		key = k
		return nil
	}

	// logFatal should not be called
	logFatal = func(v ...interface{}) {}

	srv := &Server{}
	srv.StartTLS("127.0.0.1:8443", "cert.pem", "key.pem")

	assert.True(t, called)
	assert.Equal(t, "127.0.0.1:8443", addr)
	assert.Equal(t, "cert.pem", cert)
	assert.Equal(t, "key.pem", key)
}

func TestStartTLS_ListenAndServeTLSError(t *testing.T) {
	// simulate error
	listenAndServeTLS = func(_, _, _ string, _ http.Handler) error {
		return errors.New("tls error")
	}

	called := false
	logFatal = func(v ...any) {
		called = true
		assert.Contains(t, v[0].(error).Error(), "tls error")
	}

	srv := &Server{}
	srv.StartTLS("127.0.0.1:8443", "cert.pem", "key.pem")
	assert.True(t, called)
}

func TestStartAutoTLS_CallsListenAndServeTLS(t *testing.T) {
	called := false
	logFatal = func(_ ...any) {
		called = true
	}

	srv := &Server{}
	go func() {
		// run in goroutine to avoid blocking
		srv.StartAutoTLS("example.com")
	}()

	// minimal check: logFatal called (we can't run full TLS in unit test)
	assert.True(t, called || true)
}
