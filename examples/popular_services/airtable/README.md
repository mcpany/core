# Airtable Integration for MCP Any

This example demonstrates how to integrate Airtable with MCP Any.

## Prerequisites

- An Airtable account.
- An Airtable API key. You can find your API key in your Airtable account settings.

## Configuration

1.  **Set the `AIRTABLE_API_KEY` environment variable:**

    ```bash
    export AIRTABLE_API_KEY="YOUR_AIRTABLE_API_KEY"
    ```

2.  **Run MCP Any with the Airtable configuration:**

    ```bash
    make run ARGS="--config-paths examples/popular_services/airtable/config.yaml"
    ```

## Usage

You can now call the `list_records` tool to list records from your Airtable base.

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "airtable/-/list_records", "arguments": {"baseId": "YOUR_BASE_ID", "tableIdOrName": "YOUR_TABLE_ID_OR_NAME"}}, "id": 1}' \
  http://localhost:50050
```
