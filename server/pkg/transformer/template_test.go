package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextTemplate_Render(t *testing.T) {
	t.Parallel()
	templateString := "Hello, {{name}}! You are {{age}} years old."
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{
		"name": "World",
		"age":  99,
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World! You are 99 years old.", rendered)
}

func TestTextTemplate_InvalidTemplate(t *testing.T) {
	t.Parallel()
	templateString := "Hello, {{name!"
	_, err := NewTemplate(templateString, "{{", "}}")
	require.Error(t, err)
}

func TestTextTemplate_MissingParameter(t *testing.T) {
	t.Parallel()
	templateString := "Hello, {{name}}!"
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{}
	_, err = tpl.Render(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing key")
}

func TestTextTemplate_MultiplePlaceholders(t *testing.T) {
	t.Parallel()
	templateString := "User: {{user}}, Role: {{role}}, ID: {{id}}"
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{
		"user": "admin",
		"role": "administrator",
		"id":   123,
	}
	rendered, err := tpl.Render(params)
	require.NoError(t, err)
	assert.Equal(t, "User: admin, Role: administrator, ID: 123", rendered)
}

func TestTextTemplate_CustomDelimiters(t *testing.T) {
	t.Parallel()
	templateString := "Data: [=data=], Value: [=value=]"
	tpl, err := NewTemplate(templateString, "[=", "=]")
	require.NoError(t, err)

	params := map[string]any{
		"data":  "test-data",
		"value": 456,
	}
	rendered, err := tpl.Render(params)
	require.NoError(t, err)
	assert.Equal(t, "Data: test-data, Value: 456", rendered)
}

func TestTextTemplate_EmptyTemplate(t *testing.T) {
	t.Parallel()
	templateString := ""
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{"key": "value"}
	rendered, err := tpl.Render(params)
	require.NoError(t, err)
	assert.Equal(t, "", rendered)
}

func TestTextTemplate_JSONEscaping(t *testing.T) {
	t.Parallel()
	// JSON template
	templateString := `{"key": "{{val}}"}`
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	// Malicious input
	params := map[string]any{
		"val": `", "injected": true, "ignore": "`,
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)

	// Expect escaping: " -> \"
	expected := `{"key": "\", \"injected\": true, \"ignore\": \""}`
	assert.Equal(t, expected, rendered)
}

func TestTextTemplate_JSONComplexTypes(t *testing.T) {
	t.Parallel()
	// JSON template expecting a value (not string)
	templateString := `{"config": {{config}}}`
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{
		"config": map[string]any{"a": 1},
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)

	// Expect JSON marshaling
	expected := `{"config": {"a":1}}`
	assert.Equal(t, expected, rendered)
}

func TestTextTemplate_NonJSON(t *testing.T) {
	t.Parallel()
	// Text template (no JSON escaping)
	templateString := `Value: {{val}}`
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{
		"val": `foo"bar`,
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)

	// Expect NO escaping
	expected := `Value: foo"bar`
	assert.Equal(t, expected, rendered)
}

func TestTextTemplate_FalsePositives(t *testing.T) {
	t.Parallel()

	// Case 1: Template starts with {{ (which starts with {)
	// Should NOT be detected as JSON.
	templateString := `{{val}}`
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	params := map[string]any{
		"val": `foo"bar`,
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)

	// Expect NO escaping
	assert.Equal(t, `foo"bar`, rendered)

	// Case 2: Template starts with [INFO] (starts with [)
	// Should NOT be detected as JSON.
	templateString = `[INFO] {{val}}`
	tpl, err = NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	rendered, err = tpl.Render(params)
	require.NoError(t, err)

	// Expect NO escaping
	assert.Equal(t, `[INFO] foo"bar`, rendered)
}
