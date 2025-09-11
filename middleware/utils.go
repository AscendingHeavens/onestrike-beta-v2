package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/AscendingHeavens/onestrike/v2/server"
)

// generateCSRFToken creates a random token
func generateCSRFToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// getOrCreateCSRFToken ensures a token exists in context
func getOrCreateCSRFToken(c *server.Context, cfg CSRFConfig) string {
	if t, ok := c.Params[cfg.ContextKey]; ok && t != "" {
		return t
	}
	token := generateCSRFToken(32)
	c.Params[cfg.ContextKey] = token
	return token
}

// validateCSRFToken compares HMACed tokens in constant time
func validateCSRFToken(secret []byte, serverToken, clientToken string) bool {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(serverToken))
	expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(clientToken))
}
