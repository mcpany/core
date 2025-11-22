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
