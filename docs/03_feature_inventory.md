# Feature Inventory: The Universal Agent Bus

## Status: Active & Proposed

| Priority | Feature Name | Category | Description | Status |
| :--- | :--- | :--- | :--- | :--- |
| **P0** | **Policy Firewall** | Security | Rego/CEL based hooking for tool calls. | In Development |
| **P0** | **HITL Middleware** | Safety | Suspension protocol for user approval flows. | Proposed |
| **P1** | **Recursive Context** | Comms | Standardized headers for Subagent inheritance. | In Development |
| **P1** | **Shared KV Store** | State | Embedded SQLite "Blackboard" tool for agents. | Proposed |

## New Proposals: 2026-02-22

### [P0] Isolated Docker-bound Named Pipes
*   **Description:** A new communication adapter that uses isolated named pipes within Docker containers for inter-agent communication, replacing insecure local HTTP tunneling.
*   **Target Persona:** Enterprise Local LLM Swarm Orchestrators.
*   **Strategic Value:** Directly mitigates the "OpenClaw Exploit" and enables secure agent swarms.

### [P1] Zero Trust Shell Adapter
*   **Description:** A hardened version of the Command Adapter that executes shell commands in ephemeral, resource-constrained containers with strict syscall filtering.
*   **Target Persona:** Developers using autonomous agents for local automation.

### [P1] MCP Discovery Notifications
*   **Description:** Implementation of the MCP notification protocol to allow dynamic tool discovery and updates without server restarts.
*   **Strategic Value:** Enables highly dynamic swarms where agents can "learn" about new capabilities on-the-fly.
