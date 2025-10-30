/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clock

import (
	"sync"
	"time"
)

// Clock is an interface for telling time.
type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

// New returns a new Clock that tells the current time.
func New() Clock {
	return &realClock{}
}

type realClock struct{}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func (c *realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// FakeClock is a clock that can be manually advanced.
type FakeClock struct {
	mu sync.RWMutex
	t  time.Time
}

// NewFake returns a new FakeClock that starts at the given time.
func NewFake(t time.Time) *FakeClock {
	return &FakeClock{t: t}
}

// Now returns the current time of the fake clock.
func (c *FakeClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.t
}

// Since returns the time since the given time.
func (c *FakeClock) Since(t time.Time) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.t.Sub(t)
}

// Advance advances the fake clock by the given duration.
func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}
