# Reddit

This directory contains the configuration for the Reddit upstream service.

## Tools

### `get_subreddit_info`

This tool allows you to get information about a subreddit.

#### Arguments

- `subreddit` (string): The name of the subreddit.

#### Example

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "reddit/-/get_subreddit_info", "arguments": {"subreddit": "golang"}}, "id": 1}' \
  http://localhost:50050
```
