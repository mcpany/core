// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"errors"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Retry implements a retry policy for failed operations.
type Retry struct {
	config *configv1.RetryConfig
}

// NewRetry creates a new Retry instance with the given configuration.
// It sets default values for base and max backoff if they are not provided.
// Returns the result.
func NewRetry(config *configv1.RetryConfig) *Retry {
	if config.GetBaseBackoff() == nil {
		config.SetBaseBackoff(durationpb.New(time.Second))
	}
	if config.GetMaxBackoff() == nil {
		config.SetMaxBackoff(durationpb.New(30 * time.Second))
	}
	return &Retry{
		config: config,
	}
}

// Execute runs the provided work function, retrying it if it fails according
// to the configured policy.
// Returns an error.
func (r *Retry) Execute(work func() error) error {
	var err error
	for i := 0; i < int(r.config.GetNumberOfRetries())+1; i++ {
		err = work()
		if err == nil {
			return nil
		}

		var permanentErr *PermanentError
		if errors.As(err, &permanentErr) {
			return err
		}

		time.Sleep(r.backoff(i))
	}
	return err
}

func (r *Retry) backoff(attempt int) time.Duration {
	backoff := r.config.GetBaseBackoff().AsDuration() * time.Duration(1<<attempt)
	if backoff > r.config.GetMaxBackoff().AsDuration() {
		return r.config.GetMaxBackoff().AsDuration()
	}
	return backoff
}
