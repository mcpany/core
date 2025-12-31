// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGauge_NoPanic(t *testing.T) {
	// The function signature is SetGauge(name string, val float32, labels ...string)
	// It should NOT panic if labels are empty.
	assert.NotPanics(t, func() {
		SetGauge("test_gauge", 1.0)
	}, "SetGauge should not panic if no labels are provided")

    // Test with valid labels
    assert.NotPanics(t, func() {
        SetGauge("test_gauge_with_label", 1.0, "some_service")
    }, "SetGauge should not panic with valid labels")
}
