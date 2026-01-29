package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeID_Coverage(t *testing.T) {
	// 1. len(ids) == 0
	res, err := SanitizeID([]string{}, false, 10, 8)
	assert.NoError(t, err)
	assert.Equal(t, "", res)

	// 2. Single ID, clean, no hash needed
	res, err = SanitizeID([]string{"clean"}, false, 10, 8)
	assert.NoError(t, err)
	assert.Equal(t, "clean", res)

	// 3. Single ID, empty
	res, err = SanitizeID([]string{""}, false, 10, 8)
	assert.Error(t, err)

	// 4. Dirty ID, small hash length (defaults to 8)
	// "bad!" -> "bad" + hash
	res, err = SanitizeID([]string{"bad!"}, false, 10, 0)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(res, "bad_"))
	assert.Len(t, res, 3+1+8) // 12

	// 5. Huge hash length (capped at 64)
	res, err = SanitizeID([]string{"bad!"}, false, 10, 100)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(res, "bad_"))
	assert.Len(t, res, 3+1+64) // 68

	// 6. Max sanitized prefix length limit
	longStr := strings.Repeat("a", 20)
	res, err = SanitizeID([]string{longStr}, false, 5, 8)
	assert.NoError(t, err)
	// Should truncate to 5, append hash
	assert.Equal(t, 5, len(strings.Split(res, "_")[0]))
	assert.True(t, strings.HasPrefix(res, "aaaaa_"))

	// 7. All dirty
	res, err = SanitizeID([]string{"!!!!"}, false, 5, 8)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(res, "id_"))

	// 8. Multiple IDs
	res, err = SanitizeID([]string{"a", "b"}, false, 5, 8)
	assert.NoError(t, err)
	assert.Equal(t, "a.b", res)

	// 9. Multiple IDs with one empty (should error)
	res, err = SanitizeID([]string{"a", ""}, false, 5, 8)
	assert.Error(t, err)
}

func TestReplaceURLPath_Coverage(t *testing.T) {
	// Nested braces / Unclosed braces
	params := map[string]interface{}{
		"key": "value",
	}

	// "{{key}}" -> "value"
	assert.Equal(t, "value", ReplaceURLPath("{{key}}", params, nil))

	// Let's test non-existing key
	assert.Equal(t, "{{missing}}", ReplaceURLPath("{{missing}}", params, nil))

	// Partial braces
	assert.Equal(t, "{{broken", ReplaceURLPath("{{broken", params, nil))

	// NoEscape
	params["slashed"] = "a/b"
	assert.Equal(t, "a%2Fb", ReplaceURLPath("{{slashed}}", params, nil))
	assert.Equal(t, "a/b", ReplaceURLPath("{{slashed}}", params, map[string]bool{"slashed": true}))

	// Query Escape
	assert.Equal(t, "a%2Fb", ReplaceURLQuery("{{slashed}}", params, nil))
}

func TestParseToolName_Coverage(t *testing.T) {
	s, tName, err := ParseToolName("service.tool")
	assert.NoError(t, err)
	assert.Equal(t, "service", s)
	assert.Equal(t, "tool", tName)

	s, tName, err = ParseToolName("toolOnly")
	assert.NoError(t, err)
	assert.Equal(t, "", s)
	assert.Equal(t, "toolOnly", tName)
}

func TestSanitizeOperationID_Coverage(t *testing.T) {
	// Clean
	assert.Equal(t, "clean", SanitizeOperationID("clean"))

	// Dirty
	// "bad space" -> "bad_HASH_space"
	res := SanitizeOperationID("bad space")
	assert.Contains(t, res, "bad_")
	assert.Contains(t, res, "_space")
}

func TestGetDockerCommand_Coverage(t *testing.T) {
	t.Setenv("USE_SUDO_FOR_DOCKER", "true")
	cmd, args := GetDockerCommand()
	assert.Equal(t, "sudo", cmd)
	assert.Equal(t, []string{"docker"}, args)

	t.Setenv("USE_SUDO_FOR_DOCKER", "false")
	cmd, args = GetDockerCommand()
	assert.Equal(t, "docker", cmd)
	assert.Empty(t, args)
}

func TestRandomFloat64_Coverage(t *testing.T) {
	f := RandomFloat64()
	assert.GreaterOrEqual(t, f, 0.0)
	assert.Less(t, f, 1.0)
}
