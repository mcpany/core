# Feature Inventory

## Current Features
*   **REST/HTTP Adapter:** Map REST APIs to MCP tools.
*   **gRPC Adapter:** Dynamic gRPC reflection for MCP.
*   **Command Adapter:** Execute local CLI tools as MCP tools.
*   **Filesystem Adapter:** Secure file access.
*   **Policy Engine:** Basic auth and rate limiting.

## Proposed Additions: 2026-02-22
*   **Shared Context Bus (Priority: High):** A centralized state management service allowing agents to "push" and "pull" shared context variables.
*   **MCP App UI Metadata (Priority: Medium):** Support for `ui` field in `CallToolResult` to enable iframe-based UI rendering in compatible hosts.
*   **Isolated Named Pipes (Priority: High):** For inter-agent communication on local hosts, replacing potentially insecure HTTP tunnels.
*   **DLP for UI Streams (Priority: Medium):** Scan data being sent to/from MCP App iframes for sensitive information.

## Priority Shifts
*   **Standardized Context Inheritance:** Moved from "Backlog" to "Active Development" to support growing agent swarm use cases.
