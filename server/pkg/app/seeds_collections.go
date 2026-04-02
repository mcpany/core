// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// BuiltinServiceCollections contains the official service collections.
//
// Summary: Represents a BuiltinServiceCollections.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - None.
var BuiltinServiceCollections []*configv1.Collection

func init() {
	BuiltinServiceCollections = []*configv1.Collection{
		configv1.Collection_builder{
			Name:        proto.String("Data Engineering Stack"),
			Description: proto.String("Essential tools for data pipelines (PostgreSQL, Filesystem, Python)"),
			Version:     proto.String("1.0.0"),
			Services: []*configv1.UpstreamServiceConfig{
				mkTemplate(
					"sqlite-db",
					"SQLite Database",
					`{
  "type": "object",
  "title": "SQLite Configuration",
  "properties": {
    "DB_PATH": {
      "type": "string",
      "title": "Database Path",
      "description": "Path to SQLite database file",
      "default": "./data.db"
    }
  },
  "required": ["DB_PATH"]
}`,
					"npx -y @modelcontextprotocol/server-sqlite ${DB_PATH}",
				),
			},
		}.Build(),
		configv1.Collection_builder{
			Name:        proto.String("Web Dev Assistant"),
			Description: proto.String("GitHub, Browser, and Terminal tools for web development."),
			Version:     proto.String("1.0.0"),
			Services: []*configv1.UpstreamServiceConfig{
				mkTemplate("github", "GitHub Tools", "{}", "npx -y @modelcontextprotocol/server-github"),
			},
		}.Build(),
		configv1.Collection_builder{
			Name:        proto.String("A2A Security Mesh"),
			Description: proto.String("Gold Standard demonstration of A2A Auth and secure agent-to-agent handshakes."),
			Version:     proto.String("1.0.0"),
			Services: []*configv1.UpstreamServiceConfig{
				mkTemplate(
					"a2a-auth-proxy",
					"A2A Authentication Proxy",
					`{
  "type": "object",
  "title": "A2A Proxy Configuration",
  "properties": {
    "MESH_SECRET": {
      "type": "string",
      "title": "Mesh Secret",
      "description": "Cryptographic secret for agent mesh.",
      "format": "password"
    }
  },
  "required": ["MESH_SECRET"]
}`,
					"mcpany-a2a-proxy --secret ${MESH_SECRET}",
				),
				mkTemplate(
					"a2a-status-monitor",
					"A2A Auth Status Monitor",
					`{
  "type": "object",
  "title": "Status Monitor Configuration",
  "properties": {}
}`,
					"mcpany-a2a-monitor --dashboard",
				),
			},
		}.Build(),
	}
}
