package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/Ascending-Heavens/onestrike/server"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCSRFToken_ReturnsBase64StringOfCorrectLength(t *testing.T) {
	token := generateCSRFToken(32)
	assert.NotEmpty(t, token, "token should not be empty")

	// Base64 raw URL encoding should decode back to 32 bytes
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	assert.NoError(t, err, "token should be valid base64")
	assert.Equal(t, 32, len(decoded), "decoded token should have correct byte length")
}

func TestGetOrCreateCSRFToken_CreatesAndCachesToken(t *testing.T) {
	c := &server.Context{Params: make(map[string]string)}
	cfg := CSRFConfig{ContextKey: "csrf_token"}

	// First call should create token
	token1 := getOrCreateCSRFToken(c, cfg)
	assert.NotEmpty(t, token1, "token should be generated")

	// Second call should return same token (no regeneration)
	token2 := getOrCreateCSRFToken(c, cfg)
	assert.Equal(t, token1, token2, "should reuse token from context")
}

func TestValidateCSRFToken_MatchesCorrectHMAC(t *testing.T) {
	secret := []byte("supersecret")
	serverToken := "server123"

	// Compute valid HMAC manually
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(serverToken))
	expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	assert.True(t, validateCSRFToken(secret, serverToken, expected),
		"valid HMAC must pass validation")

	assert.False(t, validateCSRFToken(secret, serverToken, "invalid"),
		"invalid client token must fail validation")

	assert.False(t, validateCSRFToken([]byte("wrongsecret"), serverToken, expected),
		"different secret must fail validation")
}
