package util

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactSecrets_URLEncoded(t *testing.T) {
	secret := "my@super@secret"
	encodedSecret := url.QueryEscape(secret) // "my%40super%40secret"

	input := "https://example.com/api?key=" + encodedSecret
	secrets := []string{secret}

	redacted := RedactSecrets(input, secrets)

	assert.NotContains(t, redacted, encodedSecret, "Encoded secret should be redacted")
	assert.Contains(t, redacted, redactedPlaceholder, "Should contain redacted placeholder")
}

func TestRedactSecrets_PathEncoded(t *testing.T) {
	secret := "secret/value with space"
	// QueryEscape turns space to +
	// PathEscape turns space to %20

	encodedPath := url.PathEscape(secret) // "secret%2Fvalue%20with%20space"

	input := "https://example.com/api/v1/" + encodedPath
	secrets := []string{secret}

	redacted := RedactSecrets(input, secrets)

	assert.NotContains(t, redacted, encodedPath, "Path encoded secret should be redacted")
	assert.Contains(t, redacted, redactedPlaceholder, "Should contain redacted placeholder")
}

func TestRedactSecrets_Mixed(t *testing.T) {
	secret := "foo bar"
	encodedQuery := "foo+bar"
	encodedPath := "foo%20bar"

	input := "Query: " + encodedQuery + ", Path: " + encodedPath
	secrets := []string{secret}

	redacted := RedactSecrets(input, secrets)

	assert.NotContains(t, redacted, encodedQuery)
	assert.NotContains(t, redacted, encodedPath)
	// It should appear twice (once for each replacement)
	assert.Equal(t, "Query: [REDACTED], Path: [REDACTED]", redacted)
}
