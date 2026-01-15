// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanForSensitiveKeys_Extra(t *testing.T) {
    // Cover the case where len(input) < 128

    input := []byte(`{"password": "123"}`)
    assert.True(t, scanForSensitiveKeys(input))

    input2 := []byte(`{"user": "alice"}`)
    assert.False(t, scanForSensitiveKeys(input2))

    // Case where second char check fails
    // "pass" vs "paaa"
    // "password" starts with 'p', next is 'a'. mask has 'a'.
    // "pb..." next is 'b'. if mask doesn't have 'b', it returns false.
    // 'p' keys: "password", "passwd", "private_key", "proxy-authorization"
    // next chars: 'a', 'r'.
    // So 'b' is not allowed.
    input3 := []byte("pbssword")
    assert.False(t, scanForSensitiveKeys(input3))
}

func TestSkipString(t *testing.T) {
    // input must start with quote
    input := []byte(`"hello" world`)
    end := skipString(input, 0)
    assert.Equal(t, 7, end) // "hello" is 7 chars

    inputEsc := []byte(`"he\"llo" world`)
    endEsc := skipString(inputEsc, 0)
    assert.Equal(t, 9, endEsc)

    inputUnclosed := []byte(`"hello`)
    endUnclosed := skipString(inputUnclosed, 0)
    assert.Equal(t, 6, endUnclosed)
}

func TestMatchFoldRest(t *testing.T) {
    // Test matchFoldRest manually
    // key is lowercase
    key := []byte("password")

    // Exact match
    assert.True(t, matchFoldRest([]byte("password"), key))
    // Upper case match
    assert.True(t, matchFoldRest([]byte("PASSWORD"), key))
    // Mixed case
    assert.True(t, matchFoldRest([]byte("PaSsWoRd"), key))
    // Mismatch
    assert.False(t, matchFoldRest([]byte("passwork"), key))
    // Short input
    assert.False(t, matchFoldRest([]byte("pass"), key))
}
