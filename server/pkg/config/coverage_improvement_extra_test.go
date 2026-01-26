package config

import (
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpand_InvalidVarNames tests edge cases for variable expansion
func TestExpand_InvalidVarNames(t *testing.T) {
	// $ followed by number or symbol should be treated as literal
	// We removed ${1var} because it is treated as a variable by handleBracedVar
	input := []byte("val: $123, $-var")
	expanded, err := expand(input)
	require.NoError(t, err)
	assert.Equal(t, "val: $123, $-var", string(expanded))
}

// TestExpand_UnclosedBrace tests unclosed brace
func TestExpand_UnclosedBrace(t *testing.T) {
	input := []byte("val: ${var")
	expanded, err := expand(input)
	require.NoError(t, err)
	assert.Equal(t, "val: ${var", string(expanded))
}

// TestResolveEnvValue_JSONList tests parsing JSON list from env var
func TestResolveEnvValue_JSONList(t *testing.T) {
	// Create a dummy message with repeated field
	// We'll use McpAnyServerConfig.ConfigPaths (repeated string)
	// Actually UpstreamServiceConfig has repeated string? No.
	// GlobalSettings has repeated string profiles.

	root := &configv1.GlobalSettings{}
	// Path to profiles
	path := []string{"profiles"}

	// JSON array of strings
	val := `["p1", "p2"]`
	res := resolveEnvValue(root, path, val)
	assert.Equal(t, []interface{}{"p1", "p2"}, res)

	// JSON array of objects (if we had repeated message)
	// Let's try GlobalSettings.Middlewares (repeated Middleware)
	root = &configv1.GlobalSettings{}
	path = []string{"middlewares"}
	val = `[{"name": "m1"}, {"name": "m2"}]`
	res = resolveEnvValue(root, path, val)

	// It should return []interface{} of maps
	list, ok := res.([]interface{})
	require.True(t, ok)
	assert.Len(t, list, 2)
	// Check content
	m1, ok := list[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "m1", m1["name"])
}

// TestResolveEnvValue_JSONList_Malformed tests malformed JSON list falls back to string/parsing
func TestResolveEnvValue_JSONList_Malformed(t *testing.T) {
	root := &configv1.GlobalSettings{}
	path := []string{"profiles"}

	// Malformed JSON
	val := `["p1", "p2"`
	res := resolveEnvValue(root, path, val)
	// Should fall back to comma splitting
	list, ok := res.([]interface{})
	require.True(t, ok)
	assert.Len(t, list, 1)
	assert.Equal(t, `["p1", "p2"`, list[0])
}

// TestWatcher_ManualTrigger tests parts of watcher logic
func TestWatcher_ManualTrigger(t *testing.T) {
	// Use NewWatcher
	w, err := NewWatcher()
	require.NoError(t, err)
	defer w.Close()

	// Watch a dir
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := tmpDir + "/config.yaml"
	err = os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)

	triggered := make(chan bool)

	// Start Watch in goroutine
	go func() {
		if err := w.Watch([]string{filePath}, func() {
			triggered <- true
		}); err != nil {
			// t.Logf("Watch error: %v", err)
		}
	}()

	// Wait a bit to ensure Watch started
	time.Sleep(100 * time.Millisecond)
}

// TestApplySetOverrides_NestedArray tests array notation in --set
func TestApplySetOverrides_NestedArray(t *testing.T) {
	m := make(map[string]interface{})
	// user.profiles[0] = "admin"
	// user.profiles[1] = "dev"
	setValues := []string{
		"users[0].id=u1",
		"users[0].profile_ids[0]=admin",
		"users[0].profile_ids[1]=dev",
	}

	root := &configv1.McpAnyServerConfig{}
	applySetOverrides(m, setValues, root)

	// Verify map structure
	// users -> "0" -> { id: u1, profile_ids: { "0": admin, "1": dev } }
	users, ok := m["users"].(map[string]interface{})
	require.True(t, ok)
	u0, ok := users["0"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "u1", u0["id"])
	pIds, ok := u0["profile_ids"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "admin", pIds["0"])
	assert.Equal(t, "dev", pIds["1"])

	// Now fixTypes should convert this map to slice
	fixTypes(m, root.ProtoReflect().Descriptor())

	usersSlice, ok := m["users"].([]interface{})
	require.True(t, ok)
	assert.Len(t, usersSlice, 1)
	u0Typed, ok := usersSlice[0].(map[string]interface{})
	require.True(t, ok)
	// Check profile_ids is slice
	pIdsSlice, ok := u0Typed["profile_ids"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, "admin", pIdsSlice[0])
	assert.Equal(t, "dev", pIdsSlice[1])
}

// TestMergeStrategies covers OneOf clearing
func TestClearOneOfSiblings(t *testing.T) {
	// UpstreamService has oneof for service type
	// If we set http_service, then grpc_service should be cleared.

	m := map[string]interface{}{
		"http_service": map[string]interface{}{"url": "http://..."},
		"grpc_service": map[string]interface{}{"addr": ":50051"},
	}

	root := &configv1.UpstreamServiceConfig{}
	// Set grpc_service initially
	m = map[string]interface{}{
		"grpc_service": map[string]interface{}{"address": ":50051"},
	}

	// Now apply http_service
	// path: ["http_service", "address"]
	// value: "http://localhost"
	applyPathToMap(m, []string{"http_service", "address"}, "http://localhost", root)

	// grpc_service should be gone
	assert.Nil(t, m["grpc_service"], "grpc_service should be cleared")
	assert.NotNil(t, m["http_service"], "http_service should be set")
}

// TestFixTypes_Recursion tests fixTypes recursion
func TestFixTypes_Recursion(t *testing.T) {
	// Nested message: UpstreamServiceConfig -> pre_call_hooks (repeated CallHook)
	// CallHook -> call_policy -> rules (repeated CallPolicyRule) -> action

	// Mock map structure
	m := map[string]interface{}{
		"pre_call_hooks": map[string]interface{}{
			"0": map[string]interface{}{
				"name": "hook1",
				"call_policy": map[string]interface{}{
					"rules": map[string]interface{}{
						"0": map[string]interface{}{
							"action": "ALLOW",
						},
					},
				},
			},
		},
	}

	root := &configv1.UpstreamServiceConfig{}
	fixTypes(m, root.ProtoReflect().Descriptor())

	hooks, ok := m["pre_call_hooks"].([]interface{})
	require.True(t, ok, "pre_call_hooks should be slice")
	assert.Len(t, hooks, 1)

	h0 := hooks[0].(map[string]interface{})
	cp := h0["call_policy"].(map[string]interface{})

	rules, ok := cp["rules"].([]interface{})
	require.True(t, ok, "rules should be slice")
	assert.Len(t, rules, 1)

	r0 := rules[0].(map[string]interface{})
	assert.Equal(t, "ALLOW", r0["action"])
}
