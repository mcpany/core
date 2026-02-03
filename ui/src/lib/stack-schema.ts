/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * The JSON Schema definition for the stack configuration file.
 *
 * This schema defines the structure and validation rules for `mcp-stack.json` files,
 * including versioning, service definitions, environment variables, and dependencies.
 * It is used for validation and to provide Intellisense in editors.
 */
export const STACK_CONFIG_SCHEMA = {
  uri: "http://mcp-any/stack-config.json",
  fileMatch: ["*"],
  schema: {
    type: "object",
    properties: {
      version: {
        type: "string",
        description: "Version of the stack configuration format",
        enum: ["1.0"]
      },
      services: {
        type: "object",
        description: "Map of service definitions",
        patternProperties: {
          "^[a-zA-Z0-9-_]+$": {
            type: "object",
            properties: {
              image: {
                type: "string",
                description: "Docker image to run"
              },
              command: {
                type: "string",
                description: "Command to execute"
              },
              working_directory: {
                type: "string",
                description: "Working directory for the process"
              },
              environment: {
                description: "Environment variables",
                oneOf: [
                  {
                    type: "object",
                    additionalProperties: { type: "string" }
                  },
                  {
                    type: "array",
                    items: { type: "string" }
                  }
                ]
              },
              ports: {
                type: "array",
                description: "List of port mappings",
                items: {
                  type: "string",
                  pattern: "^([0-9]+:)?([0-9]+)$"
                }
              },
              depends_on: {
                 type: "array",
                 description: "List of services this service depends on",
                 items: { type: "string" }
              }
            },
            additionalProperties: false
          }
        }
      }
    },
    required: ["version", "services"]
  }
};
