// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Timeout implements a timeout policy for operations.
type Timeout struct {
	duration *durationpb.Duration
}

// NewTimeout creates a new Timeout instance with the given duration.
//
// duration is the duration.
//
// Returns the result.
//
// Parameters:
//   - duration: The duration.
//
// Returns:
//   - result: The result.
//
// Side Effects:
//   - None.
func NewTimeout(duration *durationpb.Duration) *Timeout {
	return &Timeout{
		duration: duration,
	}
}

// Execute runs the provided work function with a timeout.
//
// ctx is the context for the request.
// work is the work.
//
// Returns an error if the operation fails.
//
// Parameters:
//   - ctx: The context for the operation.
//   - work: The work.
//
// Returns:
//   - result: The result.
//
// Errors:
//   - Returns error if operation fails.
//
// Side Effects:
//   - None.
func (t *Timeout) Execute(ctx context.Context, work func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, t.duration.AsDuration())
	defer cancel()
	return work(ctx)
}
