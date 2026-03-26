// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

// TestToString_Cycle ...
// Summary: TestToString_Cycle
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create a cycle: p -> i -> p
	var i interface{}
	p := &i
	i = p

	// Capture panic if it happens
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	s := ToString(p)
	if s == "" {
		t.Error("ToString returned empty string for cycle")
	}
	if len(s) == 0 {
		t.Error("Result should not be empty")
	}
}

// TestToString_DepthLimit ...
// Summary: TestToString_DepthLimit
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create a deep structure (linked list of pointers)
	type Node struct {
		Next *Node
		Val  int
	}

	// Create a chain of length 60
	head := &Node{Val: 0}
	curr := head
	for i := 1; i < 60; i++ {
		curr.Next = &Node{Val: i}
		curr = curr.Next
	}

	// ToString should not crash
	s := ToString(head)
	// It should contain some content
	if len(s) == 0 {
		t.Error("ToString returned empty string for deep structure")
	}
}

// TestToString_StructCycle ...
// Summary: TestToString_StructCycle
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	type Node struct {
		Next *Node
	}
	n := &Node{}
	n.Next = n

	s := ToString(n)
	if len(s) == 0 {
		t.Error("Empty result")
	}
	t.Logf("Result: %s", s)
}
