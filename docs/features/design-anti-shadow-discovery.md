# Design Doc: Anti-Shadow MCP Discovery Scanner

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the rise of "Easy MCP" tools like Gemini CLI and FastMCP, many developers are running local MCP servers that bypass central security policies. These "Shadow MCP Servers" create a fragmented security landscape and potential for data leakage. MCP Any needs a proactive way to discover these local servers and bring them under managed oversight.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically scan the local machine for active MCP servers (Stdio-based via process inspection and HTTP/WebSocket-based via common ports).
    *   Identify servers launched by known third-party tools (e.g., Gemini CLI, Claude Desktop).
    *   Prompt the user to "Adopt" these servers into MCP Any's managed ecosystem.
    *   Apply Zero Trust policies to adopted shadow servers.
*   **Non-Goals:**
    *   Shutting down third-party servers (unless explicitly requested by the user).
    *   Scanning the entire local network (focus is on `localhost` shadow services).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer who just installed a FastMCP tool via Gemini CLI.
*   **Primary Goal:** Bring the new FastMCP tool under MCP Any's security controls (Policy Firewall).
*   **The Happy Path (Tasks):**
    1.  User runs `fastmcp install gemini-cli my-server.py`.
    2.  MCP Any's Background Scanner detects a new listener on a common MCP port or a new `gemini-cli` process.
    3.  MCP Any UI shows a notification: "Shadow MCP Server Detected: my-server.py."
    4.  User clicks "Adopt" in the UI.
    5.  MCP Any automatically generates a configuration entry for the server and applies the default "Restricted" policy.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Engine**: Polls `lsof` / `netstat` and process lists for known MCP patterns.
    - **Identification Logic**: Matches process names or port signatures against a registry of known MCP server types.
    - **Governance Bridge**: Uses the `A2A Bridge` or `Universal Adapter` to wrap the shadow server.
*   **APIs / Interfaces:**
    - `/api/discovery/shadow-servers`: List detected unmanaged servers.
    - `/api/discovery/adopt`: Transition a shadow server to managed state.
*   **Data Storage/State:** Detected shadow servers are stored in the `Shared KV Store` with a "Pending Adoption" status.

## 5. Alternatives Considered
*   **Manual Configuration Only**: Expect users to add every server manually. *Rejected* because users prioritize speed over security, leading to shadow IT.
*   **Global Port Blocking**: Try to block all ports except those managed by MCP Any. *Rejected* as too intrusive and likely to break legitimate workflows.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Adoption is the first step to applying Zero Trust. By bringing the server into MCP Any, we can then apply the Policy Firewall and HITL middleware.
*   **Observability:** The UI dashboard will feature a "Shadow Server Map" to visualize the ratio of managed vs. unmanaged tools.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
