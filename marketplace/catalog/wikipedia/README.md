# Wikipedia Service

This service allows you to fetch articles from Wikipedia.

## Usage

To use this service, you need to provide the title of the Wikipedia page you want to fetch.

### Example

```
mcp-client wikipedia --title "Pet_door"
```

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
./build/bin/server run --config-path examples/popular_services/wikipedia/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust wikipedia http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use get_page_title_title to call get_page_title_title"
```
