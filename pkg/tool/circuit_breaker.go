/*
 * Copyright 2025 Author(s) of MCP Any
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

package tool

import (
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/sony/gobreaker"
)

var (
	circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	cbMutex         = &sync.Mutex{}
)

func GetCircuitBreaker(serviceID string, cbConfig *configv1.CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	cbMutex.Lock()
	defer cbMutex.Unlock()

	if cb, ok := circuitBreakers[serviceID]; ok {
		return cb
	}

	st := gobreaker.Settings{}
	if cbConfig != nil {
		st.Name = serviceID
		st.MaxRequests = uint32(cbConfig.GetHalfOpenRequests())
		st.Timeout = cbConfig.GetOpenDuration().AsDuration()
		st.ReadyToTrip = func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		}
	}

	cb := gobreaker.NewCircuitBreaker(st)
	circuitBreakers[serviceID] = cb
	return cb
}
