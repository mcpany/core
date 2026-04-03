// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Timeout represents the public Timeout entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Timeout struct {
	duration *durationpb.Duration
}

// NewTimeout serves as a public interface for interacting with NewTimeout.
//
// Summary: Constructs and returns an initialized timeout ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewTimeout(duration *durationpb.Duration) *Timeout {
	return &Timeout{
		duration: duration,
	}
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *Timeout) Execute(ctx context.Context, work func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, t.duration.AsDuration())
	defer cancel()
	return work(ctx)
}
