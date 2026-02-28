// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Timeout implements a timeout policy for operations.
//
// Summary: Enforces a maximum duration for operations.
type Timeout struct {
	duration *durationpb.Duration
}

// NewTimeout creates a new Timeout instance with the given duration.
//
// Summary: Initializes a new Timeout policy.
//
// Parameters:
//   - duration: *durationpb.Duration. The timeout duration.
//
// Returns:
//   - *Timeout: The initialized timeout policy.
func NewTimeout(duration *durationpb.Duration) *Timeout {
	return &Timeout{
		duration: duration,
	}
}

// Execute runs the provided work function with a timeout.
//
// Summary: Executes work within a timed context.
//
// Parameters:
//   - ctx: context.Context. The parent context.
//   - work: func(context.Context) error. The function to execute.
//
// Returns:
//   - error: An error if the work fails or the timeout is exceeded.
//
// Errors:
//   - Returns context.DeadlineExceeded if the timeout is reached.
//
// Side Effects:
//   - Creates a child context with a deadline.
func (t *Timeout) Execute(ctx context.Context, work func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, t.duration.AsDuration())
	defer cancel()
	return work(ctx)
}
