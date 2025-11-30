// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"errors"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Retry struct {
	config *configv1.RetryConfig
}

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
