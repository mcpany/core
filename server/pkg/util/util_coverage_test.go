// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"math"
	"testing"
)

func TestToString_ScientificNotation_Fix(t *testing.T) {
	// 15 digits fits in float64 without loss of precision.
	val := 123456789012345.0
	s := ToString(val)
	if s != "123456789012345" {
		t.Errorf("Expected 123456789012345, got %s", s)
	}

	// Check float32 as well
	val32 := float32(12345678.0) // 8 digits
	s = ToString(val32)
	if s != "12345678" {
		t.Errorf("Expected 12345678, got %s", s)
	}

	// Check very large number (would be scientific notation in 'g')
	// 1e20 = 100000000000000000000
	valLarge := 1e20
	s = ToString(valLarge)
	expectedLarge := "100000000000000000000"
	if s != expectedLarge {
		t.Errorf("Expected %s, got %s", expectedLarge, s)
	}

	// Check decimal places
	valDec := 1.2345
	s = ToString(valDec)
	if s != "1.2345" {
		t.Errorf("Expected 1.2345, got %s", s)
	}

	// Check small number
	valSmall := 0.000001
	s = ToString(valSmall)
	if s != "0.000001" {
		t.Errorf("Expected 0.000001, got %s", s)
	}

	// Very small number might be problematic with 'f' (lots of zeros)?
	// 1e-10 -> "0.0000000001"
	valVerySmall := 1e-10
	s = ToString(valVerySmall)
	if s != "0.0000000001" {
		t.Errorf("Expected 0.0000000001, got %s", s)
	}

	// Check NaN and Inf?
	// strconv.FormatFloat handles them.
	s = ToString(math.NaN())
	if s != "NaN" {
		t.Errorf("Expected NaN, got %s", s)
	}

	s = ToString(math.Inf(1))
	if s != "+Inf" {
		t.Errorf("Expected +Inf, got %s", s)
	}
}

func TestToString_Types(t *testing.T) {
	// Cover other types in ToString for package coverage

	// json.Number
	s := ToString(json.Number("123"))
	if s != "123" {
		t.Errorf("Expected 123, got %s", s)
	}

	// bool
	if ToString(true) != "true" { t.Error("bool true failed") }
	if ToString(false) != "false" { t.Error("bool false failed") }

	// int types
	if ToString(int(1)) != "1" { t.Error("int failed") }
	if ToString(int8(1)) != "1" { t.Error("int8 failed") }
	if ToString(int16(1)) != "1" { t.Error("int16 failed") }
	if ToString(int32(1)) != "1" { t.Error("int32 failed") }
	if ToString(int64(1)) != "1" { t.Error("int64 failed") }

	// uint types
	if ToString(uint(1)) != "1" { t.Error("uint failed") }
	if ToString(uint8(1)) != "1" { t.Error("uint8 failed") }
	if ToString(uint16(1)) != "1" { t.Error("uint16 failed") }
	if ToString(uint32(1)) != "1" { t.Error("uint32 failed") }
	if ToString(uint64(1)) != "1" { t.Error("uint64 failed") }

	// Default
	if ToString([]int{1, 2}) != "[1 2]" { t.Error("slice failed") }
}

func TestSanitizeID_Coverage(t *testing.T) {
    // SanitizeID coverage

    // empty ids
    out, err := SanitizeID(nil, false, 10, 8)
    if err != nil || out != "" {
        t.Errorf("Expected empty, nil error, got %s, %v", out, err)
    }

    // len(ids) == 1, !alwaysAppendHash
    // empty id
    out, err = SanitizeID([]string{""}, false, 10, 8)
    if err == nil {
        t.Error("Expected error for empty id")
    }

    // clean id
    out, err = SanitizeID([]string{"clean"}, false, 10, 8)
    if out != "clean" || err != nil {
        t.Errorf("Expected clean, got %s, %v", out, err)
    }

    // dirty id -> hash
    // "dirty!" -> "dirty_HASH"
    out, err = SanitizeID([]string{"dirty!"}, false, 10, 8)
    if err != nil { t.Error(err) }
    // "!" is invalid. "dirty" remains.
    // dirtyCount > 0. appendHash=true.
    // "dirty_HASH".
    // Check it starts with dirty_
    if len(out) <= 6 || out[:6] != "dirty_" {
        t.Errorf("Expected dirty_HASH..., got %s", out)
    }

    // alwaysAppendHash
    out, err = SanitizeID([]string{"clean"}, true, 10, 8)
    if len(out) <= 6 || out[:6] != "clean_" {
        t.Errorf("Expected clean_HASH..., got %s", out)
    }

    // len(ids) > 1
    // "a", "b" -> "a.b"
    out, err = SanitizeID([]string{"a", "b"}, false, 10, 8)
    if out != "a.b" {
         t.Errorf("Expected a.b, got %s", out)
    }

    // Error in loop (empty id)
    out, err = SanitizeID([]string{"a", ""}, false, 10, 8)
    if err == nil {
        t.Error("Expected error for empty id in loop")
    }

    // maxSanitizedPrefixLength
    long := "abcdefghijklmnopqrstuvwxyz"
    // max=5.
    out, err = SanitizeID([]string{long}, false, 5, 8)
    // rawSanitizedLen=26. max=5. appendHash=true.
    // finalSanitizedLen=5.
    // "abcde_HASH"
    if len(out) != 5+1+8 { // 5 chars + _ + 8 hash
        t.Errorf("Expected length %d, got %d (%s)", 14, len(out), out)
    }

    // reqHashLength <= 0
    out, err = SanitizeID([]string{"foo"}, true, 10, 0)
    // default hashLength=8.
    // "foo_HASH(8)"
    if len(out) != 3+1+8 {
         t.Errorf("Expected length 12, got %d (%s)", len(out), out)
    }

    // reqHashLength > 64
    out, err = SanitizeID([]string{"foo"}, true, 10, 100)
    // cap at 64.
    if len(out) != 3+1+64 {
        t.Errorf("Expected length 68, got %d", len(out))
    }

    // sanitizePart error?
    // Not really reachable as sanitizePart returns nil always (nolint:unparam says it returns error but it is nil).
    // The signature `func sanitizePart(...) error` is kept probably for future or consistency.
}

func TestIsNil_Coverage(t *testing.T) {
    if !IsNil(nil) { t.Error("nil should be nil") }
    var ptr *int
    if !IsNil(ptr) { t.Error("nil ptr should be nil") }
    ptr = new(int)
    if IsNil(ptr) { t.Error("non-nil ptr should not be nil") }

    // struct
    type S struct{}
    if IsNil(S{}) { t.Error("struct should not be nil") }
}

func TestGetDockerCommand_Coverage(t *testing.T) {
    // default
    cmd, args := GetDockerCommand()
    if cmd != "docker" || len(args) != 0 {
        t.Error("default docker command incorrect")
    }
}
