# Quran Cloud Service

This service provides access to the Quran Cloud API, allowing users to look up ayahs and surahs.

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
./build/bin/server run --config-path examples/popular_services/qurancloud/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust qurancloud http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use getAyah to call getAyah"
```
