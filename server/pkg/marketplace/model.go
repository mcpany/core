// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

// CommunityServer represents a community-contributed MCP server.
type CommunityServer struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	URL                 string   `json:"url"`
	Tags                []string `json:"tags"`
	Category            string   `json:"category"`
	Command             string   `json:"command"`
	ConfigurationSchema string   `json:"configurationSchema"`
}
