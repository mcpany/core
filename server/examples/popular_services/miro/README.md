
## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export MIRO_API_TOKEN=YOUR_MIRO_API_TOKEN_VALUE

./build/bin/server run --config-path examples/popular_services/miro/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust miro http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use get_boards to call get_boards"
```
