// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Timeout implements a timeout policy for operations.
//
// Summary. Enforces a maximum duration for operations.
type Timeout struct {
	duration *durationpb.Duration
}

// NewTimeout provides newtimeout functionality.
//
// Summary: NewTimeout.
//
// Parameters.
//   - duration: The parameter.
//
// Returns.
//   - result: The result.
func NewTimeout(duration *durationpb.Duration) *Timeout {
	return &Timeout{
		duration: duration,
	}
}

// Execute provides execute functionality.
//
// Summary: Execute.
//
// Parameters.
//   - ctx: The parameter.
//   - work: The parameter.
//
// Returns.
//   - result: The result.
func (t *Timeout) Execute(ctx context.Context, work func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, t.duration.AsDuration())
	defer cancel()
	return work(ctx)
}
