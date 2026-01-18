// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// Bug Fix Verification Test
func TestScanForSensitiveKeys_EscapedKey(t *testing.T) {
	// "\u0061" is 'a'. So "\u0061pi_key" is "api_key".
	// This should be detected as sensitive.
	input := []byte(`{"\u0061pi_key": "val"}`)
	sensitive := scanForSensitiveKeys(input, true)
	assert.True(t, sensitive, "Should detect escaped sensitive key when validateKeyContext is true")
}

// Edge Cases and Regression Tests
func TestRedactJSONFast_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic Redaction",
			input:    `{"token": "123"}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Malformed JSON - Unclosed String",
			input:    `{"token": "123`,
			expected: `{"token": "[REDACTED]"`, // Current behavior: redacts rest of input and stops
		},
		{
			name:     "Malformed JSON - Missing Value",
			input:    `{"token": }`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Escaped Key",
			input:    `{"\u0074oken": "123"}`,
			expected: `{"\u0074oken": "[REDACTED]"}`,
		},
		{
			name:     "False Positive Prefix",
			input:    `{"tokenizer": "123"}`,
			expected: `{"tokenizer": "123"}`,
		},
		{
			name:     "Value with comment style chars",
			input:    `{"token": // comment}`,
			expected: `{"token": "[REDACTED]" comment}`, // Current behavior: consumes // until space
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactJSON([]byte(tc.input))
			assert.Equal(t, tc.expected, string(got))
		})
	}
}

// Coverage Tests
func TestBytesToString_Coverage(t *testing.T) {
	b := []byte("hello")
	s := BytesToString(b)
	assert.Equal(t, "hello", s)

	// Ensure it handles empty
	assert.Equal(t, "", BytesToString(nil))
	assert.Equal(t, "", BytesToString([]byte{}))
}

func TestUnescapeKeySmall_Coverage(t *testing.T) {
	// Test unescapeKeySmall directly to hit edge cases
	var buf [256]byte // Larger buffer

	// Case 1: Buffer overflow
	smallBuf := make([]byte, 10)
	input := []byte("12345678901")
	_, ok := unescapeKeySmall(input, smallBuf)
	assert.False(t, ok, "Should fail on buffer overflow")

	// Case 2: Invalid escape at EOF
	input = []byte(`\`)
	_, ok = unescapeKeySmall(input, smallBuf)
	assert.False(t, ok, "Should fail on trailing backslash")

	// Case 3: Valid escapes
	input = []byte(`\b\f\n\r\t\"\\\/`)
	res, ok := unescapeKeySmall(input, buf[:])
	assert.True(t, ok)
	expected := []byte{'\b', '\f', '\n', '\r', '\t', '"', '\\', '/'}
	assert.Equal(t, expected, res)

	// Case 4: Valid unicode
	input = []byte(`\u0061`) // 'a'
	res, ok = unescapeKeySmall(input, buf[:])
	assert.True(t, ok)
	assert.Equal(t, "a", string(res))

	// Case 5: Invalid unicode (short)
	input = []byte(`\u006`)
	res, ok = unescapeKeySmall(input, buf[:])
	assert.True(t, ok)
	assert.Equal(t, "u006", string(res))

	// Case 6: Invalid unicode (bad hex)
	input = []byte(`\u006g`)
	res, ok = unescapeKeySmall(input, buf[:])
	assert.True(t, ok)
	assert.Equal(t, "u006g", string(res))

	// Case 7: Non-ASCII unicode
	input = []byte(`\u00FF`) // 255 -> replaced by '?'
	res, ok = unescapeKeySmall(input, buf[:])
	assert.True(t, ok)
	assert.Equal(t, "?", string(res))

	// Case 8: Unknown escape
	input = []byte(`\z`)
	res, ok = unescapeKeySmall(input, buf[:])
	assert.True(t, ok)
	assert.Equal(t, "z", string(res))
}

func TestIsKeyColon_Coverage(t *testing.T) {
	// isKeyColon is used in scanJSONForSensitiveKeys
	// We want to test case where it returns false (EOF without colon)

	input := []byte(`{"key"   `)
	// "key" is at index 2. Quote ends at 6.
	// isKeyColon starts at 7.
	// Scans spaces. Reaches EOF. Returns false.

	found := scanJSONForSensitiveKeys(input)
	assert.False(t, found)

	// Test case where it returns false due to non-colon char
	input = []byte(`{"key" 123`)
	found = scanJSONForSensitiveKeys(input)
	assert.False(t, found)
}

func TestResolveSecret_RemoteContent_Coverage(t *testing.T) {
	// Enable loopback for testing
	os.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_SECRETS")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") == "my-api-key" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("secret-value"))
			return
		}
		if r.Header.Get("Authorization") == "Bearer my-token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("secret-value-bearer"))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	ctx := context.Background()

	// Test 1: API Key Auth
	secret := &configv1.SecretValue{
		Value: &configv1.SecretValue_RemoteContent{
			RemoteContent: &configv1.RemoteContent{
				HttpUrl: aws.String(ts.URL),
				Auth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							ParamName: aws.String("X-Api-Key"),
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "my-api-key",
								},
							},
						},
					},
				},
			},
		},
	}

	val, err := ResolveSecret(ctx, secret)
	require.NoError(t, err)
	assert.Equal(t, "secret-value", val)

	// Test 2: Bearer Token Auth
	secret = &configv1.SecretValue{
		Value: &configv1.SecretValue_RemoteContent{
			RemoteContent: &configv1.RemoteContent{
				HttpUrl: aws.String(ts.URL),
				Auth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "my-token",
								},
							},
						},
					},
				},
			},
		},
	}

	val, err = ResolveSecret(ctx, secret)
	require.NoError(t, err)
	assert.Equal(t, "secret-value-bearer", val)

	// Test 3: Recursion Limit
	nested := &configv1.SecretValue{
		Value: &configv1.SecretValue_PlainText{
			PlainText: "base",
		},
	}

	for i := 0; i < 12; i++ {
		nested = &configv1.SecretValue{
			Value: &configv1.SecretValue_RemoteContent{
				RemoteContent: &configv1.RemoteContent{
					HttpUrl: aws.String("http://example.com"),
					Auth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_ApiKey{
							ApiKey: &configv1.APIKeyAuth{
								ParamName: aws.String("Foo"),
								Value: nested,
							},
						},
					},
				},
			},
		}
	}

	_, err = ResolveSecret(ctx, nested)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeded max recursion depth")
}
