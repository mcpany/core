# Google Calendar

This example demonstrates how to use `mcpany` to connect to the Google Calendar API to list events from a public calendar.

## Prerequisites

- An API key for the Google Calendar API. You can obtain one from the [Google Cloud Console](https://console.cloud.google.com/).

## Configuration

The configuration for this example is in `config.yaml`. It defines a single upstream service for Google Calendar.

To use this example, you need to set the `GOOGLE_API_KEY` environment variable to your Google Calendar API key.

## Usage

1. **Set the `GOOGLE_API_KEY` environment variable:**

   ```bash
   export GOOGLE_API_KEY="YOUR_GOOGLE_API_KEY"
   ```

2. **Run `mcpany` with the Google Calendar configuration:**

   ```bash
   make run ARGS="--config-paths ./examples/popular_services/google_calendar/config.yaml"
   ```

3. **Call the `list_events` tool:**

   ```bash
   curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "google_calendar/-/list_events", "arguments": {"calendarId": "en.usa#holiday@group.v.calendar.google.com"}}, "id": 1}' \
     http://localhost:50050
   ```

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export GOOGLE_API_KEY=YOUR_GOOGLE_API_KEY_VALUE

./build/bin/server run --config-path examples/popular_services/google_calendar/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust google_calendar http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use list_events to call list_events"
```
