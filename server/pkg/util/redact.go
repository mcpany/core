// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"encoding/json"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
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

	// ⚡ Bolt Optimization: Sort keys in each group by length descending.
	// This ensures that we check longer keys first (e.g. "authorization" before "auth").
	// If the longer key matches, we are done. If "auth" matched first, we might fail the boundary check
	// (e.g. next char is 'o') and then have to check "authorization" anyway.
	for i := range sensitiveKeyGroups {
		if len(sensitiveKeyGroups[i]) > 1 {
			sort.Slice(sensitiveKeyGroups[i], func(j, k int) bool {
				return len(sensitiveKeyGroups[i][j]) > len(sensitiveKeyGroups[i][k])
			})
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
//
// If the input is not valid JSON object or array, it returns the input as is.
//
// Parameters:
//   - input: []byte. The JSON input to redact.
//
// Returns:
//   - []byte: The redacted JSON output.
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
//
// Optimization: This function performs a copy-on-write.
// If no sensitive keys are found, it returns the original map (zero allocation).
// If sensitive keys are found, it returns a new map with redacted values (and copies other fields).
// Note: This aligns with RedactJSON behavior which returns original slice if clean.
//
// Parameters:
//   - m: map[string]interface{}. The map to redact.
//
// Returns:
//   - map[string]interface{}: The potentially redacted map.
func RedactMap(m map[string]interface{}) map[string]interface{} {
	redacted, changed := redactMapMaybe(m)
	if changed {
		return redacted
	}
	return m
}

// redactMapMaybe recursively checks and redacts the map.
// It returns (newMap, true) if any redaction happened.
// It returns (nil, false) if no redaction happened (optimization to avoid copying).
func redactMapMaybe(m map[string]interface{}) (map[string]interface{}, bool) {
	// First pass: check if we need to modify anything.
	// This avoids allocation for the common case (clean map).
	// We do a shallow check of keys and recursive check of values.

	// Since map iteration order is random, we can't "copy seen so far" deterministically easily
	// unless we keep track of keys or just iterate again.
	// But allocating the map when we find the first dirty key is efficient enough.
	// We will have to iterate again to copy the rest, but that's fine.

	var newMap map[string]interface{}

	// Helper to initialize newMap with a copy of m
	initMap := func() {
		if newMap == nil {
			newMap = make(map[string]interface{}, len(m))
			for mk, mv := range m {
				newMap[mk] = mv
			}
		}
	}

	for k, v := range m {
		// Check if key is sensitive
		if IsSensitiveKey(k) {
			initMap()
			// Apply redaction
			newMap[k] = redactedPlaceholder
			continue
		}

		// Check recursive values
		if nestedMap, ok := v.(map[string]interface{}); ok {
			if res, changed := redactMapMaybe(nestedMap); changed {
				initMap()
				newMap[k] = res
			}
		} else if nestedSlice, ok := v.([]interface{}); ok {
			if res, changed := redactSliceMaybe(nestedSlice); changed {
				initMap()
				newMap[k] = res
			}
		}
	}

	if newMap != nil {
		return newMap, true
	}
	return nil, false
}

func redactSliceMaybe(s []interface{}) ([]interface{}, bool) {
	var newSlice []interface{}

	for i, v := range s {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			if res, changed := redactMapMaybe(nestedMap); changed {
				if newSlice == nil {
					newSlice = make([]interface{}, len(s))
					copy(newSlice, s)
				}
				newSlice[i] = res
			}
		} else if nestedSlice, ok := v.([]interface{}); ok {
			if res, changed := redactSliceMaybe(nestedSlice); changed {
				if newSlice == nil {
					newSlice = make([]interface{}, len(s))
					copy(newSlice, s)
				}
				newSlice[i] = res
			}
		}
	}

	if newSlice != nil {
		return newSlice, true
	}
	return nil, false
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
	"passphrase", "passphrases", "ssh_key",
	"stack_trace", "stacktrace", "traceback", "exception", "error_trace",
}

// IsSensitiveKey checks if a key name suggests it contains sensitive information.
//
// Parameters:
//   - key: string. The key name to check.
//
// Returns:
//   - bool: True if the key is considered sensitive, false otherwise.
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
		slice := input[offset:]
		idx := bytes.IndexAny(slice, allSensitiveStartChars)
		if idx == -1 {
			break
		}
		matchStart := offset + idx
		// Check bounds explicitly to satisfy gosec G602
		if matchStart >= len(input) {
			break
		}
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
	j := skipWhitespaceAndComments(input, endOffset)
	if j < len(input) {
		return input[j] == ':'
	}
	return false
}

// dsnPasswordRegex handles fallback cases but we prefer net/url.
// Matches colon, followed by password (which may start with / if followed by non-/, or be empty), followed by @.
// We avoid matching :// by ensuring if it starts with /, it's not followed by another /.
// We use [^/?#\s] to stop at path/query/fragment/space, avoiding swallowing the path or matching paths as passwords.
// This also avoids matching text with spaces (e.g. "Contact: bob@example.com").
var dsnPasswordRegex = regexp.MustCompile(`(:)([^/?#@\s][^/?#\s]*|/[^/?#@\s][^/?#\s]*|)(@)`)

// dsnSchemeRegex handles fallback cases where the DSN has a scheme (://)
// We use a stricter regex that stops at whitespace, /, ?, or # to avoid swallowing
// the path or subsequent text (e.g. multiple DSNs).
// Matches scheme://user:password@
var dsnSchemeRegex = regexp.MustCompile(`(://[^/?#:\s]*):([^\s]*?)@([^/?#@\s]*)([/?#\s]|$)`)

// dsnFallbackNoAtRegex handles cases where url.Parse failed (e.g. invalid port) and there is no '@'.
// This covers "redis://:password" or "scheme://user:password" (missing host).
// It matches "://", then optional user (non-colons), then colon, then password.
// Password is terminated by /, @, whitespace, or ".
var dsnFallbackNoAtRegex = regexp.MustCompile(`(://[^:]*):([^/@\s"?]+)`)

// dsnInvalidPortRegex handles the specific Go url.Parse error message leak "invalid port".
// e.g. parse "...": invalid port ":password".
var dsnInvalidPortRegex = regexp.MustCompile(`invalid port "(:[^"]+)"`)

// RedactDSN redacts the password from a DSN string.
//
// Supported formats: postgres://user:password@host...
//
// Parameters:
//   - dsn: string. The DSN string to redact.
//
// Returns:
//   - string: The redacted DSN string.
func RedactDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err == nil && u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			u.User = url.UserPassword(u.User.Username(), redactedPlaceholder)
			// url.String() encodes the placeholder, e.g. [REDACTED] -> %5BREDACTED%5D
			// We decode it back to keep it readable, as standard loggers often prefer readable placeholders.
			// However, a simple string replace is safer than query unescaping the whole string.
			s := u.String()
			// Replace URL encoded placeholder with readable one.
			// Note: [ is %5B and ] is %5D
			// We do string replacement.
			encodedPlaceholder := "%5BREDACTED%5D"
			// Replace all occurrences of the encoded placeholder with the readable one
			return strings.Replace(s, encodedPlaceholder, redactedPlaceholder, 1)
		}
		// If parsed successfully AND found User but no password, we trust the parser.
		// This handles cases like mysql://user@host correctly (don't fallback to regex).
		return dsn
	}

	// If parsed successfully but no User info found, check for known non-DSN schemes.
	if err == nil {
		// Optimization: Check for common schemes that are NOT DSNs but might be mistaken for one
		// by the fallback regex (e.g. mailto:bob@example.com).
		// We trust url.Parse to correctly identify the scheme.
		if strings.EqualFold(u.Scheme, "mailto") {
			return dsn
		}

		// If it's a standard hierarchical URL (not Opaque) and Host is clean (no @),
		// and we didn't find User info (checked above),
		// then we assume it's safe and doesn't contain credentials.
		// We also require Host to be non-empty, because an empty Host might imply
		// the credentials are hiding in the path (e.g. "scheme:/pass@host").
		if u.Opaque == "" && u.Host != "" && !strings.Contains(u.Host, "@") {
			return dsn
		}
	}

	// Fallback to regex if parsing fails (e.g. not a valid URL)
	// OR if url.Parse succeeded but found no user/password structure (e.g. user:pass@tcp(...) where Scheme is "user" and User is nil).
	// But note: the regex is known to be imperfect for complex cases (e.g. colons in password).
	// We apply the regex as a best-effort attempt.

	// If the DSN has a scheme, use the scheme-aware regex which is more robust for complex passwords
	// (e.g. containing colons or @) but assumes a single DSN string.
	if strings.Contains(dsn, "://") {
		// Use greedy match to handle special characters in password
		dsn = dsnSchemeRegex.ReplaceAllString(dsn, "$1:"+redactedPlaceholder+"@$3$4")

		// Fallback for cases without '@' (e.g. redis://:password where url.Parse fails due to invalid port)
		if dsnFallbackNoAtRegex.MatchString(dsn) {
			dsn = dsnFallbackNoAtRegex.ReplaceAllStringFunc(dsn, func(m string) string {
				submatches := dsnFallbackNoAtRegex.FindStringSubmatch(m)
				if len(submatches) < 3 {
					return m
				}
				prefix := submatches[1]
				potentialPass := submatches[2]

				// If potential password is purely numeric, assume it is a port (e.g. http://host:8080).
				// We do not redact ports.
				isNumeric := true
				for _, r := range potentialPass {
					if !unicode.IsDigit(r) {
						isNumeric = false
						break
					}
				}

				if isNumeric && len(potentialPass) > 0 {
					return m
				}

				// Fix: http/https often use named ports (e.g. http://myservice:web) or are misinterpreted
				// as user:password when missing @. We should not redact if it looks like http/https.
				// However, if the DSN contains an '@', it strongly suggests credentials, so we should allow redaction
				// (even if dsnSchemeRegex failed to match it for some reason).
				// We only skip redaction if it's http/https AND there is no '@'.
				trimmedDSN := strings.TrimSpace(strings.ToLower(dsn))
				if (strings.HasPrefix(trimmedDSN, "http://") || strings.HasPrefix(trimmedDSN, "https://")) && !strings.Contains(dsn, "@") {
					return m
				}

				return prefix + ":" + redactedPlaceholder
			})
		}
	}

	// Handle Go url.Parse error leak "invalid port"
	dsn = dsnInvalidPortRegex.ReplaceAllString(dsn, "invalid port \":"+redactedPlaceholder+"\"")

	return dsnPasswordRegex.ReplaceAllString(dsn, "$1"+redactedPlaceholder+"$3")
}

// redactInterval represents a range of text to be redacted.
type redactInterval struct {
	start, end int
}

// RedactSecrets replaces all occurrences of the given secrets in the text with [REDACTED].
//
// Parameters:
//   - text: string. The text to redact.
//   - secrets: []string. A list of secret values to redact from the text.
//
// Returns:
//   - string: The redacted text.
func RedactSecrets(text string, secrets []string) string {
	if text == "" || len(secrets) == 0 {
		return text
	}

	// ⚡ BOLT: Optimization - Use interval merging instead of O(N) boolean mask.
	// Randomized Selection from Top 5 High-Impact Targets
	var intervals []redactInterval

	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		start := 0
		for {
			idx := strings.Index(text[start:], secret)
			if idx == -1 {
				break
			}
			absoluteIdx := start + idx
			end := absoluteIdx + len(secret)
			intervals = append(intervals, redactInterval{start: absoluteIdx, end: end})
			// Advance by len(secret) to avoid finding overlapping instances of the *same* secret
			// (e.g. "aaaa" with secret "aa" -> match at 0, next search starts at 2).
			start = end
		}
	}

	if len(intervals) == 0 {
		return text
	}

	// Sort intervals by start position
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].start < intervals[j].start
	})

	// Merge overlapping or adjacent intervals
	var merged []redactInterval
	if len(intervals) > 0 {
		current := intervals[0]
		for i := 1; i < len(intervals); i++ {
			next := intervals[i]
			// If overlaps or touches (next.start <= current.end)
			// Touching intervals (e.g. [0,3) and [3,6)) should be merged into one [0,6)
			// so they result in a single [REDACTED] placeholder.
			if next.start <= current.end {
				if next.end > current.end {
					current.end = next.end
				}
			} else {
				merged = append(merged, current)
				current = next
			}
		}
		merged = append(merged, current)
	}

	// Build result string
	var sb strings.Builder
	sb.Grow(len(text))

	lastIdx := 0
	for _, interval := range merged {
		sb.WriteString(text[lastIdx:interval.start])
		sb.WriteString(redactedPlaceholder)
		lastIdx = interval.end
	}
	sb.WriteString(text[lastIdx:])

	return sb.String()
}
