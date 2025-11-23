# Twilio

This service uses the [Twilio MCP Server](https://github.com/twilio-labs/mcp) to expose Twilio's APIs.

## Usage

This configuration automatically discovers all tools available from the Twilio MCP server.
You can use it to send messages, make calls, and interact with other Twilio resources.

```
gemini -m gemini-2.5-flash -p 'send a message to +15551234567 from +15557654321 with body "Hello, world!"'
```

## Authentication

This service requires a Twilio Account SID, API Key, and API Secret.
You can find or create your API Key and Secret in the [Twilio Console](https://www.twilio.com/docs/iam/api-keys).

You will need to set the following environment variables:

- `TWILIO_ACCOUNT_SID`
- `TWILIO_API_KEY`
- `TWILIO_API_SECRET`

The `config.yaml` file is configured to use these environment variables to authenticate with the upstream MCP server.

## Prerequisites

- **Node.js and npx**: The upstream service is a Node.js application run via `npx`. Ensure these are installed and available in your PATH.
