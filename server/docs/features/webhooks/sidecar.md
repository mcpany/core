# Webhooks Sidecar

The Webhooks Sidecar is a companion service for MCP Any that provides a set of standard, reusable webhooks for common tasks like data transformation and sanitization.

## Purpose

Instead of writing custom webhooks for every small task, you can run the sidecar alongside MCP Any and configure your upstream services to use these built-in hooks.

## Running the Sidecar

The sidecar is located in `server/cmd/webhooks`.

### Using Go

```bash
# Start the sidecar on port 8080 (default)
go run server/cmd/webhooks/main.go
```

You can customize the port using the `PORT` environment variable:

```bash
PORT=9090 go run server/cmd/webhooks/main.go
```

### Authentication

If you want to secure the sidecar, you can set a `WEBHOOK_SECRET`. The sidecar will then verify the `Webhook-Signature` header of incoming requests.

```bash
WEBHOOK_SECRET=my-secret go run server/cmd/webhooks/main.go
```

## Available Hooks

The sidecar exposes the following hooks:

### 1. Markdown Conversion (`/markdown`)

Converts HTML content in tool inputs or results to Markdown. This is useful when a tool returns raw HTML (e.g., from a web scraper) and you want to present cleaner text to the LLM.

- **Type**: Pre-Call or Post-Call
- **Endpoint**: `/markdown`
- **Configuration**:

```yaml
post_call_hooks:
  - name: "html-to-markdown"
    webhook:
      url: "http://localhost:8080/markdown"
```

### 2. Text Truncation (`/truncate`)

Truncates long strings to a specified length. Useful for preventing context window exhaustion from tools that return massive outputs.

- **Type**: Pre-Call or Post-Call
- **Endpoint**: `/truncate`
- **Parameters**: `max_chars` (query param, default: 100)
- **Configuration**:

```yaml
post_call_hooks:
  - name: "truncate-output"
    webhook:
      url: "http://localhost:8080/truncate?max_chars=500"
```

### 3. Pagination (`/paginate`)

Splits long strings into pages. This allows an LLM to request specific "pages" of data if the first response is too large, although typically this is used as a static view of the first page.

- **Type**: Pre-Call or Post-Call
- **Endpoint**: `/paginate`
- **Parameters**: `page_size` (query param, default: 1000)
- **Configuration**:

```yaml
post_call_hooks:
  - name: "paginate-output"
    webhook:
      url: "http://localhost:8080/paginate?page_size=2000"
```
