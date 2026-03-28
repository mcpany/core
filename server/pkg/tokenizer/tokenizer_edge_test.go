package tokenizer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizerCycleDetection(t *testing.T) {
	tok := NewSimpleTokenizer()

	t.Run("Map cycle", func(t *testing.T) {
		m := make(map[string]interface{})
		m["self"] = m

		_, err := CountTokensInValue(tok, m)
		assert.ErrorContains(t, err, "cycle detected in value")
	})

	t.Run("Slice cycle", func(t *testing.T) {
		s := make([]interface{}, 1)
		s[0] = s

		_, err := CountTokensInValue(tok, s)
		assert.ErrorContains(t, err, "cycle detected in value")
	})

    t.Run("Pointer cycle", func(t *testing.T) {
        type Node struct {
            Next *Node
        }
        n1 := &Node{}
        n2 := &Node{Next: n1}
        n1.Next = n2

        _, err := CountTokensInValue(tok, n1)
        assert.ErrorContains(t, err, "cycle detected in value")
    })

}

func TestReflectCountTokensCycle(t *testing.T) {
    tok := NewSimpleTokenizer()
    visited := make(map[uintptr]bool)

    m := make(map[string]interface{})
	m["self"] = m
    _, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
    assert.ErrorContains(t, err, "cycle detected in value")

    s := make([]interface{}, 1)
	s[0] = s
    _, err = countTokensReflectSlice(tok, reflect.ValueOf(s), visited)
    assert.ErrorContains(t, err, "cycle detected in value")

    type Node struct {
        Next *Node
    }
    n1 := &Node{}
    n2 := &Node{Next: n1}
    n1.Next = n2
    _, err = countTokensReflectStruct(tok, reflect.ValueOf(*n1), visited)
    assert.ErrorContains(t, err, "cycle detected in value")
}

func TestCountTokensReflectMap(t *testing.T) {
	tok := NewSimpleTokenizer()

    t.Run("Map with int key", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        m := map[int]string{1: "test"}
        count, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })

    t.Run("Map with int value", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        m := map[string]int{"test": 1}
        count, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })

    t.Run("Map with slice key", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        m := map[int]string{1: "test"}
        count, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })

    t.Run("Map with slice value", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        m := map[string][]string{"test": {"1", "2"}}
        count, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })

    t.Run("Map with map value", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        m := map[string]map[string]string{"test": {"test2": "2"}}
        count, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })

    t.Run("Map with ptr value", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        val := "2"
        m := map[string]*string{"test": &val}
        count, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })

    t.Run("Map cycle via ptr value", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        m := make(map[string]*map[string]*map[string]*string)
        m2 := make(map[string]*map[string]*string)
        m["test"] = &m2
        // cycle detected in value shouldn't crash
        _, err := countTokensReflectMap(tok, reflect.ValueOf(m), visited)
        assert.NoError(t, err)
    })

    t.Run("Slice interface simple null", func(t *testing.T) {
        visited := make(map[uintptr]bool)
        s := []interface{}{nil}
        count, err := countSliceInterfaceSimple(tok, s, visited)
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })
}
