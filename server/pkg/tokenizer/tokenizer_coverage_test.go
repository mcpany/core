package tokenizer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type failTokenizer struct {
	*SimpleTokenizer
	failOn string
}

func (t *failTokenizer) CountTokens(text string) (int, error) {
	if text == t.failOn {
		return 0, errors.New("mock error")
	}
	return t.SimpleTokenizer.CountTokens(text)
}

func (t *failTokenizer) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	if s, ok := v.(string); ok && s == t.failOn {
		return 0, errors.New("mock error")
	}
	return t.SimpleTokenizer.countRecursive(v, visited)
}

// Ensure it implements Tokenizer, not SimpleTokenizer
type genericMockTokenizer struct {
	failOn string
	*SimpleTokenizer
}

func (g *genericMockTokenizer) CountTokens(text string) (int, error) {
	if text == g.failOn {
		return 0, errors.New("mock error")
	}
	return g.SimpleTokenizer.CountTokens(text)
}

func TestTokenizerCoverage(t *testing.T) {
	st := &failTokenizer{SimpleTokenizer: NewSimpleTokenizer(), failOn: "error"}
	gt := &genericMockTokenizer{SimpleTokenizer: NewSimpleTokenizer(), failOn: "error"}

	// --- 1. reflectSlice errors ---
	sliceStr := reflect.ValueOf([]string{"ok", "error"})
	_, err := countTokensReflectSlice(st, sliceStr, make(map[uintptr]bool))
	assert.Error(t, err)

	sliceIntf := reflect.ValueOf([]interface{}{"ok", "error"})
	_, err = countTokensReflectSlice(st, sliceIntf, make(map[uintptr]bool))
	assert.Error(t, err)

	// --- 2. reflectMap errors ---
	mapStrKey := reflect.ValueOf(map[string]int{"error": 1})
	_, err = countTokensReflectMap(st, mapStrKey, make(map[uintptr]bool))
	assert.Error(t, err)

	mapIntfKey := reflect.ValueOf(map[interface{}]int{"error": 1})
	_, err = countTokensReflectMap(st, mapIntfKey, make(map[uintptr]bool))
	assert.Error(t, err)

	mapVal := reflect.ValueOf(map[int]string{1: "error"})
	_, err = countTokensReflectMap(st, mapVal, make(map[uintptr]bool))
	assert.Error(t, err)

	// --- 3. countTokensInValueRecursive errors & fallback logic ---
	// Force it to use generic fallback switch in countTokensInValueRecursive
	// By using genericMockTokenizer, it won't cast to *SimpleTokenizer or *WordTokenizer
	c, err := countTokensInValueRecursive(gt, "ok", make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	_, err = countTokensInValueRecursive(gt, "error", make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countTokensInValueRecursive(gt, []string{"ok", "error"}, make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countTokensInValueRecursive(gt, map[string]interface{}{"error": 1}, make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countTokensInValueRecursive(gt, map[string]interface{}{"ok": "error"}, make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countTokensInValueRecursive(gt, []interface{}{"ok", "error"}, make(map[uintptr]bool))
	assert.Error(t, err)

	// Hit map[string]string fallback for countTokensInValueRecursive
	c, err = countTokensInValueRecursive(gt, map[string]string{"ok": "ok2"}, make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	_, err = countTokensInValueRecursive(gt, map[string]string{"error": "ok"}, make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countTokensInValueRecursive(gt, map[string]string{"ok": "error"}, make(map[uintptr]bool))
	assert.Error(t, err)

	// Hit []int fallback for countTokensInValueRecursive
	c, err = countTokensInValueRecursive(gt, []int{1, 2}, make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	// Hit reflection fallback for countTokensInValueRecursive
	// Provide an unsupported type like channel or a struct
	type testStruct struct{ F1 string }
	c, err = countTokensInValueRecursive(gt, testStruct{F1: "ok"}, make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	_, err = countTokensInValueRecursive(gt, testStruct{F1: "error"}, make(map[uintptr]bool))
	assert.Error(t, err)

	// --- 4. countMapStringInterface & countSliceInterfaceSimple & reflectStruct errors ---
	_, err = countMapStringInterface(st, map[string]interface{}{"ok": "error"}, make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countMapStringInterface(st, map[string]interface{}{"error": 1}, make(map[uintptr]bool))
	assert.Error(t, err)

	_, err = countSliceInterfaceSimple(st.SimpleTokenizer, []interface{}{"ok", "error"}, make(map[uintptr]bool))
	// countSliceInterfaceSimple uses SimpleTokenizer internally, so we can't easily make it fail
	// unless there's a cycle. But we will cover its positive branch.
	// Actually we can test its cycle detection to get coverage on the error path.

	_, err = countTokensReflectStruct(st, reflect.ValueOf(testStruct{F1: "error"}), make(map[uintptr]bool))
	assert.Error(t, err)

	type structGen struct{ F1 interface{} }
	_, err = countTokensReflectStruct(st, reflect.ValueOf(structGen{F1: "error"}), make(map[uintptr]bool))
	assert.Error(t, err)

	// --- 5. simpleTokenizeInt/Int64 edge cases ---
	c, handled, err := countTokensInValueSimpleFast(st.SimpleTokenizer, -123)
	assert.True(t, handled)
	assert.NoError(t, err)
	assert.Equal(t, 1, c) // In SimpleTokenizer, it uses length/4. "-123" is 4 chars, 4/4 = 1.

	c, handled, err = countTokensInValueSimpleFast(st.SimpleTokenizer, int64(-123456789))
	assert.True(t, handled)
	assert.NoError(t, err)

	c, handled, err = countTokensInValueSimpleFast(st.SimpleTokenizer, 0)
	assert.True(t, handled)
	assert.Equal(t, 1, c)

	c, handled, err = countTokensInValueSimpleFast(st.SimpleTokenizer, int64(0))
	assert.True(t, handled)
	assert.Equal(t, 1, c)

	// --- 6. WordTokenizer recursive fallback ---
	wt := NewWordTokenizer()
	// To hit countTokensInValueWord generic fallback, give it a map with non-string
	c, err = countTokensInValueWord(wt, map[int]string{1: "test"}, make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	// --- 7. Generic Reflection edge cases ---
	c, err = countTokensReflectGeneric(st, "string", make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	c, err = countTokensReflect(st, reflect.ValueOf(map[int]string{1: "test"}), make(map[uintptr]bool))
	assert.NoError(t, err)
	assert.Greater(t, c, 0)

	// Pointers
	ptrInt := 1
	c, err = countTokensReflect(st, reflect.ValueOf(&ptrInt), make(map[uintptr]bool))
	assert.NoError(t, err)

	// --- 8. Cycle Detection in interfaces ---
	type cycleNode struct{ Next *cycleNode }
	node := &cycleNode{}
	node.Next = node
	_, err = countSliceInterfaceSimple(st.SimpleTokenizer, []interface{}{node}, make(map[uintptr]bool))
	assert.Error(t, err)
}
