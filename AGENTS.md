# MCP Any - Agent Instructions

## Project Goal

The primary goal of this project is to **eliminate the requirement to implement new MCP servers for doing API calls**.

Instead of writing code to create a new MCP server, users should be able to:

1.  Create MCP server tools, resources, and prompts using **only configurations**.
2.  **Share these configurations publicly**. This allows others to use the same capabilities without needing to pass around binary files or run complex setups.
3.  Run the MCP server **locally**. This ensures that sensitive data remains within the user's environment and is not exposed to remote third-party servers.

## Key Principles for Documentation and Development

- **Configuration over Code**: Emphasize that adding new capabilities is done through config files (YAML/JSON), not by writing Go/Python code for a new server.
- **Portability**: Configurations are meant to be shared. Design features and documentation to support easy sharing and importing of these configs.
- **Security & Privacy**: Highlight that the server runs locally, which is a key benefit for users concerned about leaking sensitive info to remote servers.
- **Universal Adapter**: The system acts as a universal adapter/proxy that turns existing APIs (HTTP, gRPC, OpenAPI, CLI) into MCP-compliant tools.
