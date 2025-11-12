
package tool

import (
	"testing"

	"github.com/mcpany/core/pkg/consts"
)

func TestParseToolName(t *testing.T) {
	testCases := []struct {
		name          string
		toolName      string
		wantNamespace string
		wantMethod    string
		wantErr       bool
	}{
		{
			name:          "Valid tool name",
			toolName:      "namespace" + consts.ToolNameServiceSeparator + "method",
			wantNamespace: "namespace",
			wantMethod:    "method",
			wantErr:       false,
		},
		{
			name:       "Invalid tool name - no section",
			toolName:   "namespacemethod",
			wantMethod: "namespacemethod",
			wantErr:    false,
		},
		{
			name:          "Invalid tool name - too many sections",
			toolName:      "namespace" + consts.ToolNameServiceSeparator + "method" + consts.ToolNameServiceSeparator + "extra",
			wantNamespace: "namespace",
			wantMethod:    "method" + consts.ToolNameServiceSeparator + "extra",
			wantErr:       false,
		},
		{
			name:     "Invalid tool name - empty",
			toolName: "",
			wantErr:  true,
		},
		{
			name:     "Invalid tool name - only separator",
			toolName: consts.ToolNameServiceSeparator,
			wantErr:  true,
		},
		{
			name:     "Invalid tool name - ends with separator",
			toolName: "namespace" + consts.ToolNameServiceSeparator,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namespace, method, err := ParseToolName(tc.toolName)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseToolName() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if namespace != tc.wantNamespace {
					t.Errorf("ParseToolName() namespace = %v, want %v", namespace, tc.wantNamespace)
				}
				if method != tc.wantMethod {
					t.Errorf("ParseToolName() method = %v, want %v", method, tc.wantMethod)
				}
			}
		})
	}
}
