// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package marketplace provides functionality for managing MCP service subscriptions and marketplace logic.
package marketplace

// DefaultPopularServices contains a curated list of popular MCP services used for seeding.
var DefaultPopularServices = Collection{
	Name:        "Popular MCP Services",
	Description: "A curated list of popular and useful MCP servers from the community.",
	Services: []ServiceEntry{
		{
			Name:        "Filesystem",
			Description: "Access local files safely.",
			Type:        "stdio",
			Config: map[string]string{
				"command": "npx",
				"args":    "-y @modelcontextprotocol/server-filesystem /path/to/allowed/dir",
			},
		},
		{
			Name:        "PostgreSQL",
			Description: "Read-only access to PostgreSQL databases.",
			Type:        "stdio",
			Config: map[string]string{
				"command": "npx",
				"args":    "-y @modelcontextprotocol/server-postgres postgresql://localhost/db",
			},
		},
		{
			Name:        "Brave Search",
			Description: "Search the web using Brave Search API.",
			Type:        "stdio",
			Config: map[string]string{
				"command": "npx",
				"args":    "-y @modelcontextprotocol/server-brave-search",
				"env":     "BRAVE_API_KEY=<YOUR_KEY>",
			},
		},
		{
			Name:        "Google Maps",
			Description: "Access Google Maps data.",
			Type:        "stdio",
			Config: map[string]string{
				"command": "npx",
				"args":    "-y @modelcontextprotocol/server-google-maps",
				"env":     "GOOGLE_MAPS_API_KEY=<YOUR_KEY>",
			},
		},
		{
			Name:        "Puppeteer",
			Description: "Browser automation for web scraping and testing.",
			Type:        "stdio",
			Config: map[string]string{
				"command": "npx",
				"args":    "-y @modelcontextprotocol/server-puppeteer",
			},
		},
	},
}
