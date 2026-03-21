# Open Library Service

This service provides access to the Open Library API, allowing users to look up books by ISBN and search for authors.

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
./build/bin/server run --config-path examples/popular_services/openlibrary/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust openlibrary http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use getBookByISBN to call getBookByISBN"
```
