# Design Doc: Isolated Transport (Named Pipes / Unix Sockets)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The current MCP Any implementation relies heavily on local HTTP ports for inter-agent communication. However, the discovery of the "Shadow Agent" exploit pattern has revealed that local HTTP is inherently insecure for subagent isolation. Subagents can easily scan the local network stack to find and connect to other MCP servers, bypassing the central Policy Firewall. This document proposes moving to **Isolated Transport** using Unix Named Pipes (or Domain Sockets) and Windows Named Pipes to enforce physical isolation between agents.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Eliminate local port exposure for inter-agent and agent-to-tool communication.
    *   Provide physical isolation: only the central MCP Any gateway and the specific authorized subagent can access the transport medium.
    *   Support Unix Domain Sockets (macOS/Linux) and Named Pipes (Windows).
    *   Ensure the transport is compatible with the standard MCP JSON-RPC protocol.
*   **Non-Goals:**
    *   Encrypting the transport (isolation is achieved via filesystem permissions).
    *   Supporting remote (cross-network) isolated transport (this is for local execution only).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Developer.
*   **Primary Goal:** Spawning a subagent that can only talk to the parent's tools and cannot discover other local MCP servers.
*   **The Happy Path (Tasks):**
    1.  The Developer enables `isolated_transport` in the MCP Any configuration.
    2.  MCP Any spawns a subagent and provides a unique Unix Socket path (e.g., `/tmp/mcp-any-[session-id].sock`) instead of an HTTP URL.
    3.  The subagent connects to the socket.
    4.  The subagent attempts to scan local HTTP ports for other tools but is blocked by the host's container/sandbox policy (or simply finds nothing relevant because all other tools are also on isolated sockets).
    5.  All communication is verified by the central Policy Engine before being routed.

## 4. Design & Architecture
*   **System Flow:**
    - **Session Initiation**: When a new agent session starts, MCP Any generates a cryptographically random socket path.
    - **Transport Setup**: MCP Any starts a listener on that socket.
    - **Injection**: The socket path is passed to the subagent via environment variables or a specific MCP initialization flag.
    - **Brokerage**: All messages on the socket are ingested by the `IsolatedTransportMiddleware`, validated against the `Policy Firewall`, and then forwarded to the appropriate upstream tool or agent.
*   **APIs / Interfaces:**
    - New transport type: `mcp+unix://` and `mcp+pipe://`.
*   **Data Storage/State:** Socket paths and session bindings are stored in memory and tracked in the `Multi-Agent Session Management` table.

## 5. Alternatives Considered
*   **mTLS for Local HTTP**: Encrypting and authenticating local HTTP traffic. *Rejected* due to the overhead of certificate management for ephemeral subagents and the fact that it doesn't prevent port scanning/discovery.
*   **Network Namespacing (Docker/Podman)**: Forcing all agents into isolated network namespaces. *Rejected* as the primary solution because it requires root/privileged access and is complex to manage across OSs. Named pipes offer a more portable and lightweight isolation mechanism.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Filesystem permissions on the socket/pipe are set to the strictest possible level (Owner-only).
*   **Observability:** The UI will display the transport type (e.g., "Isolated (Unix Socket)") in the session details and the `Agent Chain Tracer`.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
