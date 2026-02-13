# Jira Integration with MCP Any

This example demonstrates how to integrate Jira with MCP Any.

## Prerequisites

- A Jira account
- A Jira Personal Access Token (PAT)

## Configuration

1.  **Set the following environment variables:**

    ```bash
    export JIRA_DOMAIN="your-jira-domain.atlassian.net"
    export JIRA_USERNAME="your-jira-username"
    export JIRA_PAT="your-jira-pat"
    export JIRA_TEST_ISSUE_KEY="PROJ-123"
    export JIRA_TEST_ISSUE_SUMMARY="Test issue for MCP Any integration"
    ```

2.  **Run the MCP Any server with the Jira configuration:**

    ```bash
    make run ARGS="--config-paths ./examples/popular_services/jira/config.yaml"
    ```

## Usage

You can now use the `jira/-/get_issue` tool to get information about a Jira issue.

### Example Request

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "jira/-/get_issue", "arguments": {"issueIdOrKey": "PROJ-123"}}, "id": 1}' \
  http://localhost:50050
```

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export JIRA_DOMAIN=YOUR_JIRA_DOMAIN_VALUE
export JIRA_PAT=YOUR_JIRA_PAT_VALUE
export JIRA_USERNAME=YOUR_JIRA_USERNAME_VALUE

./build/bin/server run --config-path examples/popular_services/jira/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust jira http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "List the available tools for this service"
```
