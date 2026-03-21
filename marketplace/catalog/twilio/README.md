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

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export TWILIO_ACCOUNT_SID=YOUR_TWILIO_ACCOUNT_SID_VALUE
export TWILIO_API_KEY=YOUR_TWILIO_API_KEY_VALUE
export TWILIO_API_SECRET=YOUR_TWILIO_API_SECRET_VALUE

./build/bin/server run --config-path examples/popular_services/twilio/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust twilio http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Interact with the service"
```
