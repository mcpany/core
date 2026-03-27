// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// BuiltinServiceCollections contains the official service collections.
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
	}
}
