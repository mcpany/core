# Airtable

This directory contains the configuration for the Airtable upstream service.

## Prerequisites

To use this service, you will need to set the `AIRTABLE_API_TOKEN` environment variable to your Airtable API token.

## Tools

### `get_record`

This tool allows you to get a record from an Airtable table.

#### Arguments

- `base_id` (string): The ID of the Airtable base.
- `table_id` (string): The ID of the Airtable table.
- `record_id` (string): The ID of the record to retrieve.

#### Example

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "airtable/-/get_record", "arguments": {"base_id": "YOUR_BASE_ID", "table_id": "YOUR_TABLE_ID", "record_id": "YOUR_RECORD_ID"}}, "id": 1}' \
  http://localhost:50050
```
