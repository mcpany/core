// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// BuiltinTemplates contains the seed configurations for high-value MCP servers.
// Deprecated: Use BuiltinServiceTemplates instead. kept for TemplateManager compatibility.
var BuiltinTemplates []*configv1.UpstreamServiceConfig

// BuiltinServiceTemplates contains the rich seed configurations for service templates.
var BuiltinServiceTemplates []*configv1.ServiceTemplate

func init() {
	// Legacy templates (minimal)
	BuiltinTemplates = []*configv1.UpstreamServiceConfig{
		mkTemplate("github", "GitHub", "{}", "npx -y @modelcontextprotocol/server-github"),
	}

	// Rich templates matching UI expectations
	BuiltinServiceTemplates = []*configv1.ServiceTemplate{
		createWttrin(),
		createGoogleMaps(),
		createSlack(),
		createGithub(),
		createPostgres(),
		createFilesystem(),
	}
}

func mkTemplate(id, name, schema, command string) *configv1.UpstreamServiceConfig {
	t := &configv1.UpstreamServiceConfig{}
	t.SetId(id)
	t.SetName(name)
	t.SetVersion("1.0.0")
	t.SetConfigurationSchema(schema)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand(command)
	cmd.SetEnv(make(map[string]*configv1.SecretValue))

	t.SetCommandLineService(cmd)
	t.SetAutoDiscoverTool(true)
	return t
}

func createWttrin() *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId("wttrin")
	t.SetName("Weather (wttr.in)")
	t.SetDescription("Get real-time weather information via wttr.in.")
	t.SetIcon("cloud")
	t.SetTags([]string{"Web"})

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("weather")

	httpSvc := &configv1.HttpUpstreamService{}
	httpSvc.SetAddress("https://wttr.in")

	tool := &configv1.ToolDefinition{}
	tool.SetName("get_weather")
	tool.SetDescription("Get the weather forecast for a location.")
	tool.SetCallId("get_weather_call")
	httpSvc.SetTools([]*configv1.ToolDefinition{tool})

	call := &configv1.HttpCallDefinition{}
	call.SetEndpointPath("/{{location}}?format=j1")
	call.SetMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET)
	inputSchema := mustStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{"type": "string", "description": "City name"},
		},
		"required": []any{"location"},
	})
	call.SetInputSchema(inputSchema)

	httpSvc.SetCalls(map[string]*configv1.HttpCallDefinition{
		"get_weather_call": call,
	})

	svc.SetHttpService(httpSvc)
	t.SetServiceConfig(svc)
	return t
}

func createGoogleMaps() *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId("google-maps")
	t.SetName("Google Maps")
	t.SetDescription("Geocoding, places, and routing.")
	t.SetIcon("map")
	t.SetTags([]string{"Web"})

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("google-maps")
	svc.SetConfigurationSchema(`{
		"type": "object",
		"properties": {
			"commandLineService.env.GOOGLE_MAPS_API_KEY": {
				"type": "string",
				"title": "Google Maps API Key",
				"description": "AIza...",
				"format": "password",
				"x-key": "commandLineService.env.GOOGLE_MAPS_API_KEY"
			}
		}
	}`)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("npx -y @modelcontextprotocol/server-google-maps")

	secret := &configv1.SecretValue{}
	secret.SetPlainText("")
	cmd.SetEnv(map[string]*configv1.SecretValue{
		"GOOGLE_MAPS_API_KEY": secret,
	})

	svc.SetCommandLineService(cmd)
	t.SetServiceConfig(svc)
	return t
}

func createSlack() *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId("slack")
	t.SetName("Slack")
	t.SetDescription("Interact with Slack channels and messages.")
	t.SetIcon("message-square")
	t.SetTags([]string{"Productivity"})

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("slack")
	svc.SetConfigurationSchema(`{
		"type": "object",
		"properties": {
			"commandLineService.env.SLACK_BOT_TOKEN": {
				"type": "string",
				"title": "Slack Bot Token",
				"description": "xoxb-...",
				"format": "password",
				"x-key": "commandLineService.env.SLACK_BOT_TOKEN"
			},
			"commandLineService.env.SLACK_TEAM_ID": {
				"type": "string",
				"title": "Slack Team ID",
				"description": "T...",
				"x-key": "commandLineService.env.SLACK_TEAM_ID"
			}
		}
	}`)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("npx -y @modelcontextprotocol/server-slack")

	secretBot := &configv1.SecretValue{}
	secretBot.SetPlainText("")
	secretTeam := &configv1.SecretValue{}
	secretTeam.SetPlainText("")

	cmd.SetEnv(map[string]*configv1.SecretValue{
		"SLACK_BOT_TOKEN": secretBot,
		"SLACK_TEAM_ID":   secretTeam,
	})

	svc.SetCommandLineService(cmd)
	t.SetServiceConfig(svc)
	return t
}

func createGithub() *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId("github")
	t.SetName("GitHub")
	t.SetDescription("Integration with GitHub API.")
	t.SetIcon("github")
	t.SetTags([]string{"Dev Tools"})

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("github")
	svc.SetConfigurationSchema(`{
		"type": "object",
		"properties": {
			"commandLineService.env.GITHUB_PERSONAL_ACCESS_TOKEN": {
				"type": "string",
				"title": "GitHub Personal Access Token",
				"description": "ghp_...",
				"format": "password",
				"x-key": "commandLineService.env.GITHUB_PERSONAL_ACCESS_TOKEN"
			}
		}
	}`)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("npx -y @modelcontextprotocol/server-github")

	secret := &configv1.SecretValue{}
	secret.SetPlainText("")
	cmd.SetEnv(map[string]*configv1.SecretValue{
		"GITHUB_PERSONAL_ACCESS_TOKEN": secret,
	})

	svc.SetCommandLineService(cmd)
	t.SetServiceConfig(svc)
	return t
}

func createPostgres() *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId("postgres")
	t.SetName("PostgreSQL")
	t.SetDescription("Connect to a PostgreSQL database.")
	t.SetIcon("database")
	t.SetTags([]string{"Database"})

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("postgres-db")
	svc.SetConfigurationSchema(`{
		"type": "object",
		"properties": {
			"connectionString": {
				"type": "string",
				"title": "PostgreSQL Connection String",
				"description": "postgresql://user:password@localhost:5432/dbname",
				"x-replace-token": "{{CONNECTION_STRING}}",
				"x-key": "commandLineService.command"
			}
		}
	}`)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("npx -y @modelcontextprotocol/server-postgres {{CONNECTION_STRING}}")

	svc.SetCommandLineService(cmd)
	t.SetServiceConfig(svc)
	return t
}

func createFilesystem() *configv1.ServiceTemplate {
	t := &configv1.ServiceTemplate{}
	t.SetId("filesystem")
	t.SetName("Filesystem")
	t.SetDescription("Expose local files to the LLM.")
	t.SetIcon("file-text")
	t.SetTags([]string{"System"})

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("local-files")
	svc.SetConfigurationSchema(`{
		"type": "object",
		"properties": {
			"directories": {
				"type": "string",
				"title": "Allowed Directories",
				"description": "/path/to/folder1 /path/to/folder2",
				"x-replace-token": "{{ALLOWED_DIRECTORIES}}",
				"x-key": "commandLineService.command"
			}
		}
	}`)

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("npx -y @modelcontextprotocol/server-filesystem {{ALLOWED_DIRECTORIES}}")

	svc.SetCommandLineService(cmd)
	t.SetServiceConfig(svc)
	return t
}

func mustStruct(v map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(v)
	if err != nil {
		panic(err)
	}
	return s
}
