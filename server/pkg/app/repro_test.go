package app

import (
	"testing"
	"google.golang.org/protobuf/encoding/protojson"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestUnmarshal(t *testing.T) {
	js := `{
        "name": "test",
        "command_line_service": {
          "command": "echo",
          "resources": [
            { "uri": "test://data.json", "name": "JSON Data", "mimeType": "application/json" }
          ],
          "reads": {}
        }
      }`
	var svc configv1.UpstreamServiceConfig
	err := protojson.Unmarshal([]byte(js), &svc)
	if err != nil {
		t.Logf("Expected error: %v", err)
	} else {
		t.Error("Expected error but got nil")
	}
}
