
## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export SLACK_API_TOKEN=YOUR_SLACK_API_TOKEN_VALUE

./build/bin/server run --config-path examples/popular_services/slack/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust slack http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "List the available tools for this service"
```
