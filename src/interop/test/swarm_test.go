package test

import (
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestSwarmInterop(t *testing.T) {
	registry := interop.NewRegistry()

	registry.Register(interop.NewOpenClawAdapter())
	registry.Register(interop.NewCrewAIAdapter())
	registry.Register(interop.NewAutoGenAdapter())

	tests := []struct {
		name      string
		req       *interop.InteropRequest
		wantError bool
		wantMatch string
	}{
		{
			name: "OpenClaw to AutoGen via MCP",
			req: &interop.InteropRequest{
				Source:    interop.OpenClaw,
				Target:    interop.AutoGen,
				Protocol:  interop.MCP,
				Operation: "ping",
			},
			wantError: false,
			wantMatch: "AutoGen Executed: ping",
		},
		{
			name: "CrewAI to OpenClaw via ACP",
			req: &interop.InteropRequest{
				Source:    interop.CrewAI,
				Target:    interop.OpenClaw,
				Protocol:  interop.ACP,
				Operation: "status",
			},
			wantError: false,
			wantMatch: "OpenClaw Executed: status",
		},
		{
			name: "AutoGen to CrewAI via A2A",
			req: &interop.InteropRequest{
				Source:    interop.AutoGen,
				Target:    interop.CrewAI,
				Protocol:  interop.A2A,
				Operation: "task",
			},
			wantError: false,
			wantMatch: "CrewAI Executed: task",
		},
		{
			name: "Invalid Protocol for CrewAI (ACP not supported)",
			req: &interop.InteropRequest{
				Source:    interop.OpenClaw,
				Target:    interop.CrewAI,
				Protocol:  interop.ACP,
				Operation: "hack",
			},
			wantError: true,
			wantMatch: "protocol ACP not supported by CrewAI",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := registry.Dispatch(tc.req)

			if err != nil && !tc.wantError {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantError {
				if res.Error == nil {
					t.Fatalf("expected error but got nil")
				}
				if res.Error.Error() != tc.wantMatch {
					t.Errorf("expected error message '%s', got '%s'", tc.wantMatch, res.Error.Error())
				}
				return
			}

			if res == nil {
				t.Fatalf("expected response, got nil")
			}
			if !res.Success {
				t.Fatalf("expected success, got failure: %v", res.Error)
			}
			if res.Data != tc.wantMatch {
				t.Errorf("expected '%s', got '%v'", tc.wantMatch, res.Data)
			}
		})
	}
}
