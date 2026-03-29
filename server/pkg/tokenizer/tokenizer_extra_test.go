// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"errors"
	"reflect"
	"testing"
)

type errTokenizer struct{}

func (e errTokenizer) CountTokens(text string) (int, error) {
	return 0, errors.New("tokenization error")
}

func (e errTokenizer) CountTokensInValue(val interface{}) (int, error) {
	return 0, errors.New("tokenization error")
}

func (e errTokenizer) countRecursive(val interface{}, visited map[uintptr]bool) (int, error) {
	return 0, errors.New("tokenization error")
}

type conditionalErrTokenizer struct {
	failOnKey            string
	failOnVal            string
	failOnCountRecursive bool
}

func (e *conditionalErrTokenizer) CountTokens(text string) (int, error) {
	if text == e.failOnKey || text == e.failOnVal {
		return 0, errors.New("conditional error")
	}
	return 1, nil
}

func (e *conditionalErrTokenizer) CountTokensInValue(val interface{}) (int, error) {
	return 1, nil
}

func (e *conditionalErrTokenizer) countRecursive(val interface{}, visited map[uintptr]bool) (int, error) {
	if e.failOnCountRecursive {
		return 0, errors.New("conditional error")
	}
	return 1, nil
}

type errStringer struct{}

func (e errStringer) String() string { return "bad" }

func TestTokenizer_ReflectMap_Coverage(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	// test cycle
	m := make(map[string]interface{})
	m["self"] = m
	_, err := countTokensReflectMap(tokenizer, reflect.ValueOf(m), visited)
	if err == nil {
		t.Errorf("Expected cycle error")
	}

	visited = make(map[uintptr]bool)
	m2 := make(map[int]interface{})
	m2[1] = "test"
	_, err = countTokensReflectMap(tokenizer, reflect.ValueOf(m2), visited)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	errTok := errTokenizer{}
	visited = make(map[uintptr]bool)
	m3 := make(map[string]interface{})
	m3["key"] = 1
	_, err = countTokensReflectMap(errTok, reflect.ValueOf(m3), visited)
	if err == nil {
		t.Errorf("Expected string key error")
	}

	visited = make(map[uintptr]bool)
	m4 := make(map[int]interface{})
	m4[1] = 1
	_, err = countTokensReflectMap(errTok, reflect.ValueOf(m4), visited)
	if err == nil {
		t.Errorf("Expected non-string key error")
	}

	visited = make(map[uintptr]bool)
	m5 := make(map[int]string)
	m5[1] = "val"
	_, err = countTokensReflectMap(errTok, reflect.ValueOf(m5), visited)
	if err == nil {
		t.Errorf("Expected string value error")
	}

	visited = make(map[uintptr]bool)
	m6 := make(map[int]bool)
	m6[1] = true
	_, err = countTokensReflectMap(errTok, reflect.ValueOf(m6), visited)
	if err == nil {
		t.Errorf("Expected bool value error")
	}
}

func TestTokenizer_ReflectMap_ConditionalErrors(t *testing.T) {
	visited := make(map[uintptr]bool)

	// value string err
	c1 := &conditionalErrTokenizer{failOnVal: "badval"}
	m1 := map[int]string{1: "badval"}
	_, err := countTokensReflectMap(c1, reflect.ValueOf(m1), visited)
	if err == nil {
		t.Errorf("Expected val string error")
	}

	// value bool err
	c2 := &conditionalErrTokenizer{failOnVal: "true"}
	m2 := map[int]bool{1: true}
	_, err = countTokensReflectMap(c2, reflect.ValueOf(m2), visited)
	if err == nil {
		t.Errorf("Expected val bool error")
	}

	// value default err
	c3 := &conditionalErrTokenizer{failOnCountRecursive: true}
	m3 := map[string]int{"good": 1}
	_, err = countTokensReflectMap(c3, reflect.ValueOf(m3), visited)
	if err == nil {
		t.Errorf("Expected val default error")
	}
}

func TestTokenizer_ReflectStruct_Coverage(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	type testStruct struct {
		A string
		B bool
		C interface{}
	}

	ts := testStruct{
		A: "test",
		B: true,
		C: "other",
	}

	got, err := countTokensReflectStruct(tokenizer, reflect.ValueOf(ts), visited)
	if got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	c1 := &conditionalErrTokenizer{failOnVal: "test"}
	_, err = countTokensReflectStruct(c1, reflect.ValueOf(ts), visited)
	if err == nil {
		t.Errorf("Expected struct string error")
	}

	c2 := &conditionalErrTokenizer{failOnVal: "true"}
	_, err = countTokensReflectStruct(c2, reflect.ValueOf(ts), visited)
	if err == nil {
		t.Errorf("Expected struct bool error")
	}

	c3 := &conditionalErrTokenizer{failOnCountRecursive: true}
	_, err = countTokensReflectStruct(c3, reflect.ValueOf(ts), visited)
	if err == nil {
		t.Errorf("Expected struct generic error")
	}
}

func TestTokenizer_ReflectSlice_Coverage(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	s1 := []string{"test"}
	got, err := countTokensReflectSlice(tokenizer, reflect.ValueOf(s1), visited)
	if got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	c1 := &conditionalErrTokenizer{failOnVal: "test"}
	_, err = countTokensReflectSlice(c1, reflect.ValueOf(s1), visited)
	if err == nil {
		t.Errorf("Expected slice string error")
	}

	s2 := []int{1}
	c3 := &conditionalErrTokenizer{failOnCountRecursive: true}
	_, err = countTokensReflectSlice(c3, reflect.ValueOf(s2), visited)
	if err == nil {
		t.Errorf("Expected slice generic error")
	}
}

func TestTokenizer_Reflect_Coverage(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	type testStruct struct {
		A string
	}

	ts := &testStruct{A: "test"}
	got, err := countTokensReflect(tokenizer, ts, visited)
	if got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// cycle
	type Node struct {
		Next *Node
	}
	n1 := &Node{}
	n2 := &Node{Next: n1}
	n1.Next = n2
	_, err = countTokensReflect(tokenizer, n1, visited)
	if err == nil {
		t.Errorf("Expected cycle error")
	}

	// slice cycle
	s := make([]interface{}, 1)
	s[0] = s
	_, err = countTokensReflect(tokenizer, s, visited)
	if err == nil {
		t.Errorf("Expected slice cycle error")
	}

	// struct cycle via ptr
	type Tree struct {
		Left *Tree
	}
	t1 := &Tree{}
	t1.Left = t1
	_, err = countTokensReflect(tokenizer, t1, visited)
	if err == nil {
		t.Errorf("Expected struct cycle error")
	}

	// Array (fallback)
	arr := [1]int{1}
	_, err = countTokensReflect(tokenizer, arr, visited)
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}

	// Map cycle
	m := make(map[string]interface{})
	m["self"] = m
	_, err = countTokensReflect(tokenizer, m, visited)
	if err == nil {
		t.Errorf("Expected map cycle error")
	}

	// stringer error
	c1 := &conditionalErrTokenizer{failOnVal: "bad"}
	_, err = countTokensReflect(c1, errStringer{}, visited)
	if err == nil {
		t.Errorf("Expected error from countTokensReflect stringer")
	}

	// fallback for unknown type
	ch := make(chan int)
	_, err = countTokensReflect(tokenizer, ch, visited)
	if err != nil {
		t.Errorf("Unexpected err for channel: %v", err)
	}
}

func TestSimpleTokenizeInt64_Extra(t *testing.T) {
	// hit < 1 branch
	// n < 0, l=1
	// n = -n -> 1000000
	// switch: < 10000000 -> l += 7 -> l = 8
	// count = 8/4 = 2.

	// Wait, if n = -1000000, l=1, n=1000000, l+=7 -> l=8, 8/4 = 2.
	// If count < 1 is to be hit, l < 4.
	// But fast path handles > -1000000 and < 10000000.
	// So any negative <= -1000000 has l >= 1 + 7 = 8.
	// Any positive >= 10000000 has l >= 8.
	// So l is ALWAYS >= 8.
	// Thus count = l/4 is ALWAYS >= 2.
	// Thus `if count < 1 { return 1 }` is unreachable!
	// But we just test a few values to make sure.
	tests := []int64{
		-1000000,
		-9999999,
		10000000,
		99999999,
	}

	for _, v := range tests {
		res := simpleTokenizeInt64(v)
		if res < 1 {
			t.Errorf("expected > 0, got %d for %d", res, v)
		}
	}
}

func TestSimpleTokenizer_SliceErrors(t *testing.T) {
	st := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	// Test cycle in countSliceInterfaceSimple
	s := make([]interface{}, 1)
	s[0] = s
	_, err := countSliceInterfaceSimple(st, s, visited)
	if err == nil {
		t.Errorf("Expected cycle error in countSliceInterfaceSimple")
	}
}

// Extra coverage for WordTokenizer
func TestWordTokenizer_CountTokens(t *testing.T) {
	wt := NewWordTokenizer()

	// Coverage for CountTokens fallback logic.
	_, err := wt.CountTokens("simple text")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// struct cycle via ptr
	type Tree struct {
		Left *Tree
	}
	t1 := &Tree{}
	t1.Left = t1
	visited := make(map[uintptr]bool)
	_, err = countTokensReflect(wt, t1, visited)
	if err == nil {
		t.Errorf("Expected struct cycle error")
	}

	// Map cycle for word tokenizer
	visited = make(map[uintptr]bool)
	m := make(map[string]interface{})
	m["self"] = m
	_, err = countTokensReflect(wt, m, visited)
	if err == nil {
		t.Errorf("Expected map cycle error")
	}

	// Slice cycle for word tokenizer
	visited = make(map[uintptr]bool)
	s := make([]interface{}, 1)
	s[0] = s
	_, err = countTokensReflect(wt, s, visited)
	if err == nil {
		t.Errorf("Expected slice cycle error")
	}

	visited = make(map[uintptr]bool)
	m2 := make(map[int]interface{})
	m2[1] = "test"
	got, err := countTokensReflect(wt, reflect.ValueOf(m2), visited)
	if err != nil {
		t.Errorf("Unexpected error for map: %v", err)
	}
	if got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestCycleDetection_AllTypes(t *testing.T) {
	st := NewSimpleTokenizer()
	wt := NewWordTokenizer()

	// pointer cycle
	visited := make(map[uintptr]bool)

	i := 5
	p := &i
	_, err := countTokensReflectGeneric(st, p, visited)
	if err != nil {
		t.Errorf("Unexpected pointer error: %v", err)
	}

	// nil pointer
	var nilPtr *int
	_, err = countTokensReflectGeneric(st, nilPtr, make(map[uintptr]bool))
	if err != nil {
		t.Errorf("Unexpected nil pointer error: %v", err)
	}

	// ptr cycle already covered in TestTokenizer_Reflect_Coverage via t1.Left = t1

	// struct containing map
	visited = make(map[uintptr]bool)
	type mapStruct struct {
		M map[string]interface{}
	}
	m := make(map[string]interface{})
	ms := mapStruct{M: m}
	m["self"] = ms
	_, err = countTokensReflectGeneric(st, ms, visited)
	// struct is pass-by-value, but map is reference. The cycle is in map.
	if err == nil {
		t.Errorf("Expected cycle error in mapStruct")
	}

	// struct containing slice
	visited = make(map[uintptr]bool)
	type sliceStruct struct {
		S []interface{}
	}
	s := make([]interface{}, 1)
	ss := sliceStruct{S: s}
	s[0] = ss
	_, err = countTokensReflectGeneric(wt, ss, visited)
	if err == nil {
		t.Errorf("Expected cycle error in sliceStruct")
	}
}

func TestSimpleTokenizeInt64_Edges(t *testing.T) {
	tests := []struct {
		in   int64
		want int
	}{
		{0, 1},
		{10, 1},
		{100, 1},
		{1000, 1},
		{10000, 1},
		{100000, 1},
		{1000000, 1},
		{10000000, 2},
		{100000000, 2},
		{1000000000, 2},
		{10000000000, 2},
		{100000000000, 3},
		{1000000000000, 3},
		{10000000000000, 3},
		{100000000000000, 3},
		{1000000000000000, 4},
		{10000000000000000, 4},
		{100000000000000000, 4},
		{1000000000000000000, 4},
		{-1000000000000000000, 5},
	}

	for _, tt := range tests {
		got := simpleTokenizeInt64(tt.in)
		if got != tt.want {
			t.Errorf("simpleTokenizeInt64(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestSliceInterfaceSimple_Extra(t *testing.T) {
	st := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	s := []interface{}{
		map[string]interface{}{"key": 123},
		map[int]interface{}{1: "test"}, // hits default
		[]interface{}{"nested"},
	}

	got, err := countSliceInterfaceSimple(st, s, visited)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if got != 5 {
		t.Errorf("expected 5, got %d", got)
	}

}

func TestSliceInterfaceSimple_CyclesInside(t *testing.T) {
	st := NewSimpleTokenizer()

	// Create cycle within map inside slice
	s := make([]interface{}, 1)
	m := make(map[string]interface{})
	m["self"] = m
	s[0] = m

	_, err := countSliceInterfaceSimple(st, s, make(map[uintptr]bool))
	if err == nil {
		t.Errorf("Expected cycle error in map inside slice")
	}

	// Create cycle within slice inside slice
	s2 := make([]interface{}, 1)
	s3 := make([]interface{}, 1)
	s3[0] = s3
	s2[0] = s3

	_, err = countSliceInterfaceSimple(st, s2, make(map[uintptr]bool))
	if err == nil {
		t.Errorf("Expected cycle error in slice inside slice")
	}

	// Cycle in default path (e.g. pointer)
	s4 := make([]interface{}, 1)

	// ptr cycle not easy without recursive struct, let's use a struct with pointer to self
	type PtrNode struct {
		Next *PtrNode
	}
	n := &PtrNode{}
	n.Next = n
	s4[0] = n

	_, err = countSliceInterfaceSimple(st, s4, make(map[uintptr]bool))
	if err == nil {
		t.Errorf("Expected cycle error in default (struct ptr) inside slice")
	}
}
