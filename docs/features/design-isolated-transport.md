# Design Doc: Isolated Transport (Unix Domain Sockets / Named Pipes)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The current MCP ecosystem relies heavily on local HTTP servers for inter-process communication (IPC) between agents and tools. However, the rise of "Shadow Agent" exploits—where rogue subagents or compromised scripts scan local ports to bypass parent-level restrictions—has exposed a critical vulnerability in the local HTTP model. MCP Any must provide a more secure, isolated transport mechanism that does not rely on discoverable network ports.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement Unix Domain Sockets (UDS) and Named Pipes (Windows) as first-class transport layers in MCP Any.
    *   Enable agents and tools to communicate via isolated file-based sockets with strict filesystem permissions.
    *   Maintain backward compatibility with standard JSON-RPC over MCP.
    *   Provide a secure "Socket Registry" to replace port-based discovery.
*   **Non-Goals:**
    *   Replacing HTTP for remote (cross-network) MCP communication.
    *   Implementing a new protocol (this is a transport-layer enhancement).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Power User / Security Researcher.
*   **Primary Goal:** Execute a local file-system tool via an agent without exposing any local network ports that could be scanned by other running processes or subagents.
*   **The Happy Path (Tasks):**
    1.  The user configures an MCP tool to use a Unix socket: `mcpany --socket /tmp/mcp-fs.sock`.
    2.  MCP Any starts the tool and listens on the specified socket file.
    3.  The agent (e.g., Claude Code) connects to the tool via the socket instead of `http://localhost:8080`.
    4.  Filesystem permissions (chmod 600) ensure only the authorized agent process can access the socket.
    5.  The tool call is executed, and no local ports are ever opened.

## 4. Design & Architecture
*   **System Flow:**
    - **Transport Layer**: The `IsolatedTransport` module abstracts the connection logic. It detects the OS and selects either UDS (Linux/macOS) or Named Pipes (Windows).
    - **Security Handshake**: Upon connection, MCP Any verifies the UID/GID of the connecting process (on supported systems) to ensure it matches the authorized agent's identity.
    - **Multiplexing**: Support for multiple concurrent agent sessions over a single socket file using framing.
*   **APIs / Interfaces:**
    - `listen(socket_path string)`
    - `dial(socket_path string)`
    - Integration with existing `MCPTransport` interface.
*   **Data Storage/State:** Socket paths and authorized process IDs are tracked in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Loopback-Only HTTP (127.0.0.1)**: Already implemented, but still vulnerable to port scanning by any process running on the same machine. *Rejected* for high-security use cases.
*   **Shared Memory**: Extremely fast but complex to implement across different languages and runtimes. *Rejected* in favor of the more standard socket-based IPC.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Socket files should be created in a restricted directory with `700` permissions. Use `SO_PEERCRED` on Linux to verify the peer's credentials.
*   **Observability:** Log socket lifecycle events (Creation, Connection, Termination) in the Audit Log. Provide a "Socket Health" indicator in the UI.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
