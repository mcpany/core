# Design Doc: Non-TCP Transport (UDS/Named Pipes)

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
The "ClawJacked" (CVE-2026-25253) exploit demonstrated that local TCP-based agent gateways are vulnerable to browser-based cross-origin attacks. Malicious websites can open WebSocket connections to `localhost` and brute-force credentials. To eliminate this entire attack surface, MCP Any needs a transport layer that is inaccessible to the browser. Unix Domain Sockets (UDS) and Windows Named Pipes provide a robust solution by moving inter-process communication (IPC) out of the network stack and into the filesystem, where standard web browsers have no reach.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement support for Unix Domain Sockets (UDS) on Linux/macOS and Named Pipes on Windows.
    *   Provide a standard way for MCP clients and servers to connect via these non-TCP transports.
    *   Ensure that the `mcpany` CLI and SDKs prioritize UDS/Named Pipes for local connections when available.
    *   Implement file-system-based permissions for UDS/Named Pipe access, providing an additional layer of security.
*   **Non-Goals:**
    *   Replacing TCP entirely (it is still needed for remote or containerized environments where UDS isn't practical).
    *   Implementing a new protocol (this is a transport-layer change only).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious developer running a local agent swarm.
*   **Primary Goal:** Connect multiple local agents to the MCP Any gateway without exposing any TCP ports.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any with `--transport uds`.
    2.  MCP Any creates a socket file at `/tmp/mcpany.sock` (with restricted permissions).
    3.  User configures their agents to use the UDS path instead of a localhost URL.
    4.  Agents connect securely; browser-based malicious scripts cannot see or interact with the `/tmp/mcpany.sock` file.

## 4. Design & Architecture
*   **System Flow:**
    - **Transport Abstraction**: The `mcpserver` package will be updated to accept a generic `net.Listener`.
    - **UDS Listener**: A new listener implementation that handles socket file creation, cleanup, and permission setting (`chmod 600`).
    - **Client Discovery**: The discovery service will include the UDS path in the metadata for local services.
*   **APIs / Interfaces:**
    - CLI Flag: `--uds-path [path]`
    - SDK Connection String: `unix:///tmp/mcpany.sock` or `np:\\.\pipe\mcpany`
*   **Data Storage/State:** The socket file exists only while the process is running. MCP Any must ensure clean removal on shutdown.

## 5. Alternatives Considered
*   **Custom Browser Extension**: Building an extension to block localhost WebSockets. *Rejected* as it requires user installation and doesn't solve the problem for all browsers/agents.
*   **Strict Token-Based Auth on TCP**: Always requiring a token for localhost. *Rejected* as it still leaves the port open for discovery and brute-force attempts; UDS is fundamentally more secure for IPC.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** UDS allows for OS-level identity verification (e.g., `SO_PEERCRED`), enabling MCP Any to know exactly which PID/User is connecting.
*   **Observability:** Logs should clearly indicate which transport is being used for each active connection.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
