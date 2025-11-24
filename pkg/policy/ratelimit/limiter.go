// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"golang.org/x/time/rate"
)

// Limiter defines the interface for a rate limiter.
type Limiter interface {
	// Allow reports whether an event may happen now.
	Allow() bool
}

// inMemoryLimiter is an in-memory implementation of the Limiter interface.
type inMemoryLimiter struct {
	limiter *rate.Limiter
}

// NewInMemoryLimiter creates a new in-memory rate limiter.
func NewInMemoryLimiter(requestsPerSecond float64, burst int) Limiter {
	return &inMemoryLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
	}
}

// Allow checks if a request is allowed based on the rate limit.
func (l *inMemoryLimiter) Allow() bool {
	return l.limiter.Allow()
}
