// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Summary: Implements timeout logic.
//
// Timeout implements a timeout policy for operations.
type Timeout struct {
	duration *durationpb.Duration
}

// Summary: Initializes a new Timeout policy.
//
// NewTimeout creates a new Timeout instance with the given duration.
//
// Parameters:
//   - duration: *durationpb.Duration. The duration of the timeout.
//
// Returns:
//   - *Timeout: The initialized timeout policy.
func NewTimeout(duration *durationpb.Duration) *Timeout {
	return &Timeout{
		duration: duration,
	}
}

// Summary: Executes a function with a timeout.
//
// Execute runs the provided work function with a timeout.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - work: func(context.Context) error. The function to execute.
//
// Returns:
//   - error: An error if the operation fails or times out.
//
// Throws/Errors:
//   - context.DeadlineExceeded: If the operation times out.
func (t *Timeout) Execute(ctx context.Context, work func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, t.duration.AsDuration())
	defer cancel()
	return work(ctx)
}
