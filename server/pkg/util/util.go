// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/consts"
)

// SanitizeID sanitizes a slice of strings to form a valid ID.
// It performs the following operations:
//
//  1. Joining the strings with a "." separator.
//
//  2. Removing any characters that are not allowed (alphanumerics, "_", "-").
//     Allowed characters are: `[a-zA-Z0-9_-]`.
//
//  3. Truncating the result to the specified maximum length.
//
//  4. Optionally, appending a hash of the original string to ensure uniqueness,
//     especially when truncation occurs or when illegal characters are present.
//
// After sanitizing each string individually, it joins them with a "." separator to form
// the final identifier.
//
// Summary: Creates a sanitized, unique identifier from a list of string parts.
//
// Parameters:
//   - ids: []string. A slice of strings to be sanitized and joined.
//   - alwaysAppendHash: bool. If true, forces a hash to be appended.
//   - maxSanitizedPrefixLength: int. The maximum allowed length for the sanitized prefix.
//   - reqHashLength: int. The desired length of the hexadecimal hash to be appended.
//
// Returns:
//   - string: A single string representing the sanitized and joined identifier.
//   - error: An error if the sanitization fails (e.g., if ids is empty).
func SanitizeID(ids []string, alwaysAppendHash bool, maxSanitizedPrefixLength, reqHashLength int) (string, error) {
	if len(ids) == 0 {
		return "", nil
	}

	// Optimization: Fast path for single ID that is already valid and doesn't need hashing.
	if len(ids) == 1 && !alwaysAppendHash {
		id := ids[0]
		if id == "" {
			return "", fmt.Errorf("id cannot be empty")
		}
		if len(id) <= maxSanitizedPrefixLength {
			isClean := true
			for i := 0; i < len(id); i++ {
				if !allowedSanitizeIDChars[id[i]] {
					isClean = false
					break
				}
			}
			if isClean {
				return id, nil
			}
		}
	}

	// Determine effective hash length for total length calculation
	effectiveHashLength := reqHashLength
	if effectiveHashLength <= 0 {
		effectiveHashLength = hashLength // Default from package constant (8)
	} else if effectiveHashLength > 64 {
		effectiveHashLength = 64 // Cap at SHA256 hex length
	}

	// Optimization: Pre-calculate total length to avoid multiple allocations
	totalLen := 0
	for i, id := range ids {
		if id == "" {
			return "", fmt.Errorf("id cannot be empty")
		}
		totalLen += len(id)
		// Optimization: Always reserve space for the hash to avoid reallocation if the ID turns out to be dirty.
		// Over-allocating by ~9 bytes per ID is better than re-allocating the entire buffer.
		totalLen += 1 + effectiveHashLength // _ + hash
		if i > 0 {
			totalLen++ // dot
		}
	}

	var sb strings.Builder
	sb.Grow(totalLen)

	for i, id := range ids {
		if i > 0 {
			sb.WriteByte('.')
		}

		if err := sanitizePart(&sb, id, alwaysAppendHash, maxSanitizedPrefixLength, reqHashLength); err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

func sanitizePart(sb *strings.Builder, id string, alwaysAppendHash bool, maxSanitizedPrefixLength, reqHashLength int) error { //nolint:unparam
	// Check quick conditions for mandatory hashing
	if alwaysAppendHash || len(id) > maxSanitizedPrefixLength {
		// Must append hash.
		// Write up to maxSanitizedPrefixLength clean chars.
		written := 0
		for i := 0; i < len(id) && written < maxSanitizedPrefixLength; i++ {
			if isValidChar(id[i]) {
				sb.WriteByte(id[i])
				written++
			}
		}

		if written == 0 {
			sb.WriteString("id")
		}

		appendHashSuffix(sb, id, reqHashLength)
		return nil
	}

	// Here, len(id) <= maxSanitizedPrefixLength AND !alwaysAppendHash.
	// We only append hash if we find a dirty char.

	// Scan for dirty char
	dirtyIdx := -1
	for i := 0; i < len(id); i++ {
		if !isValidChar(id[i]) {
			dirtyIdx = i
			break
		}
	}

	if dirtyIdx == -1 {
		// All clean, no hash needed.
		sb.WriteString(id)
		return nil
	}

	// Found dirty char at dirtyIdx.
	// We must append hash.

	// We write everything before dirtyIdx (all clean).
	sb.WriteString(id[:dirtyIdx])
	written := dirtyIdx

	// Continue from dirtyIdx + 1, skipping dirty chars.
	for i := dirtyIdx + 1; i < len(id); i++ {
		if isValidChar(id[i]) {
			sb.WriteByte(id[i])
			written++
		}
	}

	if written == 0 {
		sb.WriteString("id")
	}

	appendHashSuffix(sb, id, reqHashLength)
	return nil
}

func appendHashSuffix(sb *strings.Builder, id string, reqHashLength int) {
	// Append Hash
	// Optimization: Use sha256.Sum256 to avoid heap allocation of hash.Hash
	sum := sha256.Sum256(stringToBytes(id))

	// Avoid hex.EncodeToString allocation
	var hashBuf [64]byte // sha256 hex is 64 chars
	hex.Encode(hashBuf[:], sum[:])

	sb.WriteByte('_')

	// Determine effective hash length
	effectiveLen := reqHashLength
	if effectiveLen <= 0 {
		effectiveLen = hashLength // Use package-level constant (8)
	} else if effectiveLen > 64 {
		effectiveLen = 64
	}

	sb.Write(hashBuf[:effectiveLen])
}

func isValidChar(c byte) bool {
	return allowedSanitizeIDChars[c]
}

// SanitizeServiceName sanitizes the given service name.
// It ensures that the name is a valid identifier by removing disallowed characters
// and appending a hash if the name is too long or contains illegal characters.
// This function calls SanitizeID with alwaysAppendHash set to false.
//
// Summary: Sanitizes a service name to be safe for use as an identifier.
//
// Parameters:
//   - name: string. The service name to sanitize.
//
// Returns:
//   - string: The sanitized service name.
//   - error: An error if sanitization fails.
func SanitizeServiceName(name string) (string, error) {
	return SanitizeID([]string{name}, false, maxSanitizedPrefixLength, hashLength)
}

// SanitizeToolName sanitizes the given tool name.
// It ensures that the name is a valid identifier by removing disallowed characters
// and appending a hash if the name is too long or contains illegal characters.
// This function calls SanitizeID with alwaysAppendHash set to false.
//
// Summary: Sanitizes a tool name to be safe for use as an identifier.
//
// Parameters:
//   - name: string. The tool name to sanitize.
//
// Returns:
//   - string: The sanitized tool name.
//   - error: An error if sanitization fails.
func SanitizeToolName(name string) (string, error) {
	return SanitizeID([]string{name}, false, maxSanitizedPrefixLength, hashLength)
}

const (
	maxSanitizedPrefixLength = 53
	hashLength               = 8
	maxGeneratedIDLength     = maxSanitizedPrefixLength + 1 + hashLength
)

var (
	// allowedIDChars is a lookup table for allowed characters in an operation ID.
	// It corresponds to the regex `[a-zA-Z0-9-._~:/?#\[\]@!$&'()*+,;=]`.
	allowedIDChars [256]bool

	// allowedSanitizeIDChars is a lookup table for allowed characters in a sanitized ID.
	// It corresponds to the regex `[a-zA-Z0-9_-]`.
	allowedSanitizeIDChars [256]bool
)

func init() {
	const allowed = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" +
		"-._~:/?#[]@!$&'()*+,;="
	for i := 0; i < len(allowed); i++ {
		allowedIDChars[allowed[i]] = true
	}

	const allowedSanitize = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" +
		"_-"
	for i := 0; i < len(allowedSanitize); i++ {
		allowedSanitizeIDChars[allowedSanitize[i]] = true
	}
}

// TrueStr is a string constant for "true".
const TrueStr = "true"

// GenerateUUID creates a new random (version 4) UUID.
//
// Summary: Generates a random UUID.
//
// Returns:
//   - string: A string representation of the UUID (e.g., "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx").
func GenerateUUID() string {
	return uuid.New().String()
}

// ParseToolName deconstructs a fully qualified tool name into its service key
// and bare tool name components. It splits the name using the standard
// separator.
//
// Summary: Parses a fully qualified tool name.
//
// Parameters:
//   - toolName: string. The fully qualified tool name to parse.
//
// Returns:
//   - service: string. The service key part.
//   - bareToolName: string. The tool name part.
//   - err: error. An error if parsing fails.
func ParseToolName(toolName string) (service, bareToolName string, err error) {
	parts := strings.SplitN(toolName, consts.ToolNameServiceSeparator, 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", toolName, nil
}

// SanitizeOperationID cleans an input string to make it suitable for use as an
// operation ID. It replaces any sequence of disallowed characters with a short
// hexadecimal hash of that sequence, ensuring uniqueness while preserving as
// much of the original string as possible.
//
// Summary: Sanitizes a string for use as an operation ID.
//
// Parameters:
//   - input: string. The string to be sanitized.
//
// Returns:
//   - string: The sanitized string.
func SanitizeOperationID(input string) string {
	// Fast path: check if valid without allocating
	isClean := true
	for i := 0; i < len(input); i++ {
		if !allowedIDChars[input[i]] {
			isClean = false
			break
		}
	}

	if isClean {
		return input
	}

	// Optimization: Calculate exact required size to avoid buffer resizing
	needed := 0
	for i := 0; i < len(input); {
		if allowedIDChars[input[i]] {
			needed++
			i++
		} else {
			// Found start of disallowed sequence
			for i < len(input) && !allowedIDChars[input[i]] {
				i++
			}
			needed += 8 // _ + 6 hash + _
		}
	}

	var sb strings.Builder
	sb.Grow(needed)

	for i := 0; i < len(input); {
		if allowedIDChars[input[i]] {
			sb.WriteByte(input[i])
			i++
		} else {
			// Found start of disallowed sequence
			start := i
			for i < len(input) && !allowedIDChars[input[i]] {
				i++
			}
			// Disallowed sequence is input[start:i]
			badChunk := input[start:i]

			// Optimization: Use sha256.Sum256 to avoid heap allocation of hash.Hash
			// Use zero-copy string to byte conversion
			sum := sha256.Sum256(stringToBytes(badChunk))
			var hashBuf [64]byte
			hex.Encode(hashBuf[:], sum[:])

			sb.WriteString("_")
			sb.Write(hashBuf[:6])
			sb.WriteString("_")
		}
	}
	return sb.String()
}

// stringToBytes converts a string to a byte slice without allocation.
// IMPORTANT: The returned byte slice must not be modified.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s)) //nolint:gosec // Standard zero-copy conversion
}

// BytesToString converts a byte slice to a string without allocation.
// IMPORTANT: The byte slice must not be modified while the string is in use.
//
// Summary: Zero-copy conversion from bytes to string.
//
// Parameters:
//   - b: []byte. The byte slice to convert.
//
// Returns:
//   - string: The resulting string.
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b)) //nolint:gosec // Standard zero-copy conversion
}

// GetDockerCommand returns the command and base arguments for running Docker.
// It checks the USE_SUDO_FOR_DOCKER environment variable to determine if
// "sudo" should be prepended to the command.
//
// Summary: Retrieves the appropriate Docker command (with optional sudo).
//
// Returns:
//   - string: The command to run (e.g., "docker" or "sudo").
//   - []string: The arguments for the command (e.g., [] or ["docker"]).
//
// Side Effects:
//   - Reads environment variable USE_SUDO_FOR_DOCKER.
func GetDockerCommand() (string, []string) {
	const dockerCmd = "docker"
	if os.Getenv("USE_SUDO_FOR_DOCKER") == TrueStr {
		return "sudo", []string{dockerCmd}
	}
	return dockerCmd, []string{}
}

// ReplaceURLPath replaces placeholders in a URL path with values from a params map.
// It handles URL escaping of values unless specified otherwise.
//
// Summary: Substitutes placeholders in a URL path.
//
// Parameters:
//   - urlPath: string. The URL path containing placeholders in the format "{{key}}".
//   - params: map[string]interface{}. A map of keys to values to replace placeholders with.
//   - noEscapeParams: map[string]bool. A map of keys that should NOT be URL escaped.
//
// Returns:
//   - string: The URL path with placeholders replaced.
func ReplaceURLPath(urlPath string, params map[string]interface{}, noEscapeParams map[string]bool) string {
	return replacePlaceholders(urlPath, params, noEscapeParams, url.PathEscape)
}

// ReplaceURLQuery replaces placeholders in a URL query string with values from a params map.
// It handles URL query escaping of values unless specified otherwise.
//
// Summary: Substitutes placeholders in a URL query string.
//
// Parameters:
//   - urlQuery: string. The URL query string containing placeholders in the format "{{key}}".
//   - params: map[string]interface{}. A map of keys to values to replace placeholders with.
//   - noEscapeParams: map[string]bool. A map of keys that should NOT be URL escaped.
//
// Returns:
//   - string: The URL query string with placeholders replaced.
func ReplaceURLQuery(urlQuery string, params map[string]interface{}, noEscapeParams map[string]bool) string {
	return replacePlaceholders(urlQuery, params, noEscapeParams, url.QueryEscape)
}

func replacePlaceholders(input string, params map[string]interface{}, noEscapeParams map[string]bool, escapeFunc func(string) string) string {
	var sb strings.Builder
	// Heuristic: grow slightly more than original to accommodate values
	sb.Grow(len(input) + 32)

	start := 0
	for {
		idx := strings.Index(input[start:], "{{")
		if idx == -1 {
			sb.WriteString(input[start:])
			break
		}
		absoluteIdx := start + idx
		sb.WriteString(input[start:absoluteIdx])

		end := strings.Index(input[absoluteIdx+2:], "}}")
		if end == -1 {
			sb.WriteString(input[absoluteIdx:])
			break
		}
		absoluteEnd := absoluteIdx + 2 + end

		key := input[absoluteIdx+2 : absoluteEnd]
		v, ok := params[key]
		if !ok {
			sb.WriteString(input[absoluteIdx : absoluteEnd+2])
		} else {
			val := ToString(v)
			if noEscapeParams == nil || !noEscapeParams[key] {
				val = escapeFunc(val)
			}
			sb.WriteString(val)
		}
		start = absoluteEnd + 2
	}
	return sb.String()
}

// IsNil checks if an interface value is nil or holds a nil pointer.
//
// Summary: Checks if a value is nil or a nil interface.
//
// Parameters:
//   - i: any. The interface value to check.
//
// Returns:
//   - bool: True if the value is nil or a nil pointer, false otherwise.
func IsNil(i any) bool {
	if i == nil {
		return true
	}
	// Optimization: Use reflect.ValueOf directly which contains the type information.
	// This avoids an extra reflect.TypeOf call.
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice, reflect.Func, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}

// ToString converts a value to a string representation efficiently.
// It handles common types like string, json.Number, int, float, and bool
// without using reflection when possible.
// Optimization: We manually handle all standard Go numeric types to avoid the overhead
// of reflection (fmt.Sprintf) which is significantly slower and generates more allocations.
//
// Summary: Converts any value to a string efficiently.
//
// Parameters:
//   - v: any. The value to convert to a string.
//
// Returns:
//   - string: The string representation of the value.
func ToString(v any) string {
	return toStringRecursive(v, 0)
}

func toStringRecursive(v any, depth int) string {
	// Prevent infinite recursion due to cycles
	if depth > 50 {
		return "<recursion limit reached>"
	}

	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case error:
		return val.Error()
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return float32ToString(val)
	case float64:
		return float64ToString(val)
	case fmt.Stringer:
		return val.String()
	default:
		// Check for pointer type
		rVal := reflect.ValueOf(v)
		if rVal.Kind() == reflect.Ptr {
			if rVal.IsNil() {
				return "<nil>"
			}
			return toStringRecursive(rVal.Elem().Interface(), depth+1)
		}
		return fmt.Sprintf("%v", v)
	}
}

// ⚡ BOLT: Helper to format float32 avoiding scientific notation for large values
// Randomized Selection from Top 5 High-Impact Targets (Maintenance for Fix)
func float32ToString(val float32) string {
	// Check if it's an integer and within safe range for exact representation.
	// float32 has 23 bits of significand, so exact integers up to 2^24 (16,777,216).
	if val == float32(int32(val)) {
		return strconv.FormatInt(int64(val), 10)
	}
	// Also check if it fits in int64 (for larger integers that are exact in float32)
	val64 := float64(val)
	if math.Trunc(val64) == val64 {
		if val64 >= float64(math.MinInt64) && val64 < float64(math.MaxInt64) {
			return strconv.FormatInt(int64(val), 10)
		}
	}
	// Avoid scientific notation for values within reasonable range (e.g. +/- 1e15).
	// This ensures IDs like 3000000000 don't become 3e+09.
	// We only apply this to large numbers (> 1e6) to preserve scientific notation for very small numbers.
	// We cap at 1e15 to ensure very large numbers (like MaxInt64 ~9e18) still use scientific notation
	// as expected by other tests (e.g. TestToString_BugFix).
	absVal := math.Abs(val64)
	if absVal >= 1e6 && absVal < 1e15 {
		return strconv.FormatFloat(val64, 'f', -1, 32)
	}
	return strconv.FormatFloat(val64, 'g', -1, 32)
}

// ⚡ BOLT: Helper to format float64 avoiding scientific notation for large values
// Randomized Selection from Top 5 High-Impact Targets (Maintenance for Fix)
func float64ToString(val float64) string {
	// Check if it's an integer and within int64 range.
	// float64 has 53 bits of significand. int64 is 64 bits.
	// We only convert if it fits in int64 and is an exact integer.
	// math.MinInt64 and math.MaxInt64 are boundaries.
	// However, casting large float to int64 is undefined if it overflows.
	// Safe integer range for float64 is +/- 2^53. MaxInt64 is 2^63-1.
	// So any safe float64 integer fits in int64.
	// We check if the float value is integral.
	if math.Trunc(val) == val {
		// Check bounds to avoid undefined behavior or overflow when casting.
		// float64 can exactly represent integers up to 2^53.
		// Beyond that, it can represent some integers but with gaps.
		// We only use decimal formatting if it is reasonably representative of the value.
		if val >= float64(math.MinInt64) && val < float64(math.MaxInt64) {
			return strconv.FormatInt(int64(val), 10)
		}
	}
	// Avoid scientific notation for values within reasonable range (e.g. +/- 1e15).
	// We only apply this to large numbers (> 1e6) to preserve scientific notation for very small numbers.
	// We cap at 1e15 to ensure very large numbers (like MaxInt64 ~9e18) still use scientific notation.
	absVal := math.Abs(val)
	if absVal >= 1e6 && absVal < 1e15 {
		return strconv.FormatFloat(val, 'f', -1, 64)
	}
	return strconv.FormatFloat(val, 'g', -1, 64)
}

// RandomFloat64 returns a random float64 in [0.0, 1.0).
// It uses the global math/rand source.
//
// Summary: Generates a random float.
//
// Returns:
//   - float64: A random float64 value.
func RandomFloat64() float64 {
	return rand.Float64() //nolint:gosec // Weak random is sufficient for jitter
}

// SanitizeFilename cleans a filename to ensure it is safe to use.
// It removes any directory components, null bytes, and restricts characters
// to alphanumeric, dots, dashes, and underscores.
//
// Summary: Sanitizes a filename.
//
// Parameters:
//   - filename: string. The filename to sanitize.
//
// Returns:
//   - string: The sanitized filename.
func SanitizeFilename(filename string) string {
	// 1. Base name only
	filename = filepath.Base(filename)

	// 2. Remove any null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// 3. Remove non-allowed characters
	var sb strings.Builder
	for _, c := range filename {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '.' || c == '-' || c == '_' {
			sb.WriteRune(c)
		} else {
			sb.WriteRune('_')
		}
	}
	result := sb.String()

	// 4. Ensure not empty
	if result == "" || result == "." || result == ".." {
		return "unnamed_file"
	}

	// 5. Truncate
	if len(result) > 255 {
		result = result[:255]
		// Ensure we don't end with an incomplete multi-byte character
		for len(result) > 0 {
			r, size := utf8.DecodeLastRuneInString(result)
			if r == utf8.RuneError && size == 1 {
				result = result[:len(result)-1]
			} else {
				break
			}
		}
	}

	return result
}
