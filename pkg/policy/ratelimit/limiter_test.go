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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryLimiter(t *testing.T) {
	limiter := NewInMemoryLimiter(10, 1)

	// Allow the first request
	assert.True(t, limiter.Allow(), "First request should be allowed")

	// Exhaust the burst capacity
	for i := 0; i < 10; i++ {
		limiter.Allow()
	}

	// The next request should be denied
	assert.False(t, limiter.Allow(), "Request after exhausting burst should be denied")

	// Wait for the limiter to refill
	time.Sleep(100 * time.Millisecond)

	// The next request should be allowed
	assert.True(t, limiter.Allow(), "Request after refill should be allowed")
}
