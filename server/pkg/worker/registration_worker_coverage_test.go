// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"errors"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistrationWorker_Start_BusErrors(t *testing.T) {
	tests := []struct {
		name      string
		failTopic string
	}{
		{
			name:      "fail_request_bus",
			failTopic: bus.ServiceRegistrationRequestTopic,
		},
		{
			name:      "fail_result_bus",
			failTopic: bus.ServiceRegistrationResultTopic,
		},
		{
			name:      "fail_list_request_bus",
			failTopic: bus.ServiceListRequestTopic,
		},
		{
			name:      "fail_list_result_bus",
			failTopic: bus.ServiceListResultTopic,
		},
		{
			name:      "fail_get_request_bus",
			failTopic: bus.ServiceGetRequestTopic,
		},
		{
			name:      "fail_get_result_bus",
			failTopic: bus.ServiceGetResultTopic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalTestLock.Lock()
			defer globalTestLock.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			messageBus := bus_pb.MessageBus_builder{}.Build()
			messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
			bp, err := bus.NewProvider(messageBus)
			require.NoError(t, err)

			// Save previous hook
			prevHook := bus.GetBusHook
			bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
				if topic == tt.failTopic {
					return nil, errors.New("simulated bus error")
				}
				return nil, nil
			}
			t.Cleanup(func() {
				bus.GetBusHook = prevHook
			})

			registry := &mockServiceRegistry{}
			worker := NewServiceRegistrationWorker(bp, registry)

			// This should log an error and return, not panic.
			worker.Start(ctx)
		})
	}
}
