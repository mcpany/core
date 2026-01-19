// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"encoding/json"
	"unsafe"
)

const redactedPlaceholder = "[REDACTED]"

var (
	sensitiveKeysBytes [][]byte
	redactedValue      json.RawMessage

	// sensitiveStartChars contains the lowercase starting characters of all sensitive keys.
	// Used for optimized scanning.
	sensitiveStartChars []byte

	// sensitiveKeyGroups maps a starting character (lowercase) to the list of sensitive keys starting with it.
	// Optimization: Use array instead of map for faster lookup.
	sensitiveKeyGroups [256][][]byte

	// sensitiveNextCharMask maps a starting character to a bitmask of allowed second characters.
	// Bit 0 = 'a', Bit 1 = 'b', etc.
	// Used to quickly filter out false positives based on the second character.
	sensitiveNextCharMask [256]uint32

	// sensitiveStartCharBitmap is a bitmap for fast checking if a character is a start char.
	// It's faster than bytes.IndexAny for short strings because it avoids overhead.
	sensitiveStartCharBitmap [256]bool

	// allSensitiveStartChars is a string containing all characters that can start a sensitive key.
	// Used for optimized scanning with bytes.IndexAny.
	allSensitiveStartChars string
)

func init() {
	for _, k := range sensitiveKeys {
		kb := []byte(k)
		sensitiveKeysBytes = append(sensitiveKeysBytes, kb)

		if len(kb) > 0 {
			first := kb[0] // sensitiveKeys are lowercase
			if len(sensitiveKeyGroups[first]) == 0 {
				sensitiveStartChars = append(sensitiveStartChars, first)
			}
			sensitiveKeyGroups[first] = append(sensitiveKeyGroups[first], kb)
		}
	}

	// Build sensitiveStartCharsAny and bitmap
	for _, c := range sensitiveStartChars {
		sensitiveStartCharBitmap[c] = true
		// Add uppercase variant
		upper := c - 32
		sensitiveStartCharBitmap[upper] = true
	}

	// Build allSensitiveStartChars
	var sb bytes.Buffer
	for c := 0; c < 256; c++ {
		if sensitiveStartCharBitmap[c] {
			sb.WriteByte(byte(c))
		}
	}
	allSensitiveStartChars = sb.String()

	// Build next char masks
	for start, keys := range sensitiveKeyGroups {
		if len(keys) == 0 {
			continue
		}
		var mask uint32
		for _, k := range keys {
			if len(k) > 1 {
				second := k[1] // k is lowercase
				if second >= 'a' && second <= 'z' {
					mask |= 1 << (second - 'a')
				}
			}
		}
		sensitiveNextCharMask[start] = mask
	}

	// Pre-marshal the redacted placeholder to ensure valid JSON and avoid repeated work.
	b, _ := json.Marshal(redactedPlaceholder)
	redactedValue = json.RawMessage(b)
}

// RedactJSON parses a JSON byte slice and redacts sensitive keys.
// If the input is not valid JSON object or array, it returns the input as is.
func RedactJSON(input []byte) []byte {
	// Check if input looks like JSON object or array.
	// We skip whitespace and comments to find the first significant character.
	idx := skipWhitespaceAndComments(input, 0)
	if idx >= len(input) {
		return input
	}
	first := input[idx]
	if first != '{' && first != '[' {
		return input
	}

	// Use fast zero-allocation redaction path
	// This avoids expensive json.Unmarshal/Marshal for large payloads
	return redactJSONFast(input)
}

// RedactMap recursively redacts sensitive keys in a map.
// Note: This function creates a deep copy of the map with redacted values.
func RedactMap(m map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{}, len(m))
	for k, v := range m {
		if IsSensitiveKey(k) {
			newMap[k] = redactedPlaceholder
		} else {
			if nestedMap, ok := v.(map[string]interface{}); ok {
				newMap[k] = RedactMap(nestedMap)
			} else if nestedSlice, ok := v.([]interface{}); ok {
				newMap[k] = redactSlice(nestedSlice)
			} else {
				newMap[k] = v
			}
		}
	}
	return newMap
}

func redactSlice(s []interface{}) []interface{} {
	newSlice := make([]interface{}, len(s))
	for i, v := range s {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			newSlice[i] = RedactMap(nestedMap)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			newSlice[i] = redactSlice(nestedSlice)
		} else {
			newSlice[i] = v
		}
	}
	return newSlice
}

// bytesContainsFold2 is a proposed optimization that we might use in the future.
// Ideally, we want a function that can search for multiple keys at once (Aho-Corasick),
// but for now we stick to optimizing the single key search or the calling pattern.

// sensitiveKeys is a list of substrings that suggest a key contains sensitive information.
// Note: Shorter keys that are substrings of longer keys (e.g. "token" vs "access_token") cover the longer cases,
// so we only include the shorter ones to optimize performance.
var sensitiveKeys = []string{
	"api_key", "apikey", "token", "secret", "password", "passwd", "credential", "auth", "private_key",
	"authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key",
	"passwords", "tokens", "api_keys", "apikeys",
	"authentication", "authenticator", "credentials", "secrets",
}

// IsSensitiveKey checks if a key name suggests it contains sensitive information.
//
// key is the key.
//
// Returns true if successful.
func IsSensitiveKey(key string) bool {
	// Use the optimized byte-based scanner for keys as well.
	// Avoid allocation using zero-copy conversion.
	//nolint:gosec // Zero-copy conversion for optimization
	// For IsSensitiveKey, we don't need to validate the key context (quotes and colon) because we are checking the key itself.
	return scanForSensitiveKeys(unsafe.Slice(unsafe.StringData(key), len(key)), false)
}

// scanForSensitiveKeys checks if input contains any sensitive key.
// If validateKeyContext is true, it checks if the match is followed by a closing quote and a colon.
// This function replaces the old linear scan (O(N*M)) with a more optimized scan
// that uses SIMD-accelerated IndexByte for grouped start characters.
func scanForSensitiveKeys(input []byte, validateKeyContext bool) bool { //nolint:unparam
	// Optimization: If we are validating key context (JSON input), we can scan by quotes.
	// This allows us to skip scanning string values entirely, which is a huge win for large payloads.
	if validateKeyContext {
		return scanJSONForSensitiveKeys(input)
	}

	// Optimization: For short strings, IndexAny is faster (one pass).
	// For long strings, multiple IndexByte calls are faster (SIMD).
	// The crossover is around 128 bytes.
	if len(input) < 128 {
		// Use bitmap for faster check than IndexAny on short strings
		for i := 0; i < len(input); i++ {
			c := input[i]
			if sensitiveStartCharBitmap[c] {
				startChar := c | 0x20 // Normalize to lowercase
				if checkPotentialMatch(input, i, startChar) {
					return true
				}
			}
		}
		return false
	}

	// Optimization: Use IndexAny to find the first occurrence of ANY sensitive start char.
	// This reduces the number of passes over the data from N (number of unique start chars) to 1.
	offset := 0
	for offset < len(input) {
		idx := bytes.IndexAny(input[offset:], allSensitiveStartChars)
		if idx == -1 {
			break
		}
		matchStart := offset + idx
		c := input[matchStart]

		// We found a character 'c' which is a start char.
		// We need to know which 'startChar' (lowercase) it corresponds to.
		lowerC := c | 0x20 // Normalize to lowercase

		// Check if it matches
		if checkPotentialMatch(input, matchStart, lowerC) {
			return true
		}

		offset = matchStart + 1
	}
	return false
}

// checkPotentialMatch checks if a sensitive key starts at matchStart.
// startChar must be the lowercase version of input[matchStart].
func checkPotentialMatch(input []byte, matchStart int, startChar byte) bool {
	// Optimization: Check second character
	if matchStart+1 < len(input) {
		second := input[matchStart+1] | 0x20
		if second >= 'a' && second <= 'z' {
			mask := sensitiveNextCharMask[startChar]
			if (mask & (1 << (second - 'a'))) == 0 {
				// Second character doesn't match any key in this group
				return false
			}
		}
	} else {
		// Not enough bytes for any key
		return false
	}

	keys := sensitiveKeyGroups[startChar]

	// Check all keys in this group against input starting at matchStart
	for _, key := range keys {
		if matchFoldRest(input[matchStart:], key) {
			endIdx := matchStart + len(key)
			// Check boundary: if the next character is a lowercase letter,
			// it's likely a continuation of a word (e.g. "auth" in "author"), so we skip it.
			// We allow uppercase letters (CamelCase) and other characters (snake_case, end of string).
			if endIdx < len(input) {
				next := input[endIdx]
				if next >= 'a' && next <= 'z' {
					continue
				}
				// Special handling for uppercase keys (e.g. "AUTH" in "AUTHORITY")
				// If the matched key was uppercase, and the next char is uppercase, it's a continuation.
				// However, if the matched key was lowercase (e.g. "auth" in "authToken"), it's CamelCase (boundary).
				if next >= 'A' && next <= 'Z' {
					// Check if the matched key was uppercase.
					// We know input[matchStart] matched the key start.
					// If input[matchStart] is uppercase, assume the whole key match was uppercase (or case-insensitive matching logic holds).
					firstChar := input[matchStart]
					if firstChar >= 'A' && firstChar <= 'Z' {
					// Verify if the REST of the key is also uppercase.
					// If not (e.g. "Auth"), then it is PascalCase and SHOULD be redacted.
					isAllUpper := true
					for k := 1; k < len(key); k++ {
						c := input[matchStart+k]
						if c >= 'a' && c <= 'z' {
							isAllUpper = false
							break
						}
					}
					if isAllUpper {
						continue
					}
					}
				}
			}

			return true
		}
	}
	return false
}

// matchFoldRest checks if s starts with key (case-insensitive).
// It assumes the first character already matched (case-insensitive).
func matchFoldRest(s, key []byte) bool {
	if len(s) < len(key) {
		return false
	}
	// Skip index 0 as it was already matched
	for i := 1; i < len(key); i++ {
		c := s[i]
		k := key[i] // k is lowercase

		if c == k {
			continue
		}

		// Optimization: case-insensitive comparison
		// We know k is lowercase [a-z] or special char [0-9_-].
		// If k is [a-z], (c | 0x20) == k handles both c==k and c==upper(k).
		// If k is not [a-z], we must check c == k.
		if k >= 'a' && k <= 'z' {
			if (c | 0x20) != k {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// scanJSONForSensitiveKeys scans a JSON input for sensitive keys by navigating quotes.
// It assumes keys are enclosed in quotes.
func scanJSONForSensitiveKeys(input []byte) bool {
	i := 0
	for i < len(input) {
		idx := bytes.IndexByte(input[i:], '"')
		if idx == -1 {
			return false
		}
		start := i + idx

		// We found a string starting at start.
		// Get the end of the string.
		end := skipString(input, start) // returns index after closing quote

		// Check if it is a key (followed by colon)
		if isKeyColon(input, end) {
			// It is a key. Check content.
			// Content is input[start+1 : end-1]
			if start+1 < end-1 {
				keyContent := input[start+1 : end-1]
				// Check if keyContent is sensitive.
				// We use scanForSensitiveKeys with validateKeyContext=false to check the text.
				if scanForSensitiveKeys(keyContent, false) {
					return true
				}
			}
		}

		// Move past this string
		i = end
	}
	return false
}

// isKeyColon checks if the JSON element ending at endOffset is followed by a colon.
func isKeyColon(input []byte, endOffset int) bool {
	for j := endOffset; j < len(input); j++ {
		c := input[j]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		return c == ':'
	}
	return false
}
