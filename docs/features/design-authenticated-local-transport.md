# Design Doc: Authenticated Local Transport (Named Pipes/UDS)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The "ClawJacked" vulnerability (CVE-2026-25253) demonstrated that unauthenticated local HTTP listeners are a significant security risk for AI agents. Any website or local process can potentially send requests to an agent's local server, leading to unauthorized tool execution or data exfiltration. MCP Any must transition to transports that provide OS-level authentication and isolation.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Deprecate unauthenticated local HTTP endpoints.
    *   Implement Unix Domain Sockets (UDS) with UID/GID verification for Linux/macOS.
    *   Implement Named Pipes with strictly defined Security Descriptors for Windows.
    *   Ensure that only the user who started the MCP Any process (or authorized groups) can connect to the transport.
*   **Non-Goals:**
    *   Implementing a custom application-layer authentication protocol (the OS handles it at the transport layer).
    *   Deprecating remote HTTP (Remote HTTP will still use mTLS/Tokens).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local AI Developer.
*   **Primary Goal:** Securely connect an agent (e.g., Claude Code or a custom script) to the local MCP Any gateway without exposing a network port.
*   **The Happy Path (Tasks):**
    1.  User starts `mcpany-server`.
    2.  Server creates a socket file at `/tmp/mcpany.sock` (on macOS) with `0600` permissions.
    3.  User configures their agent to use the UDS transport instead of `http://localhost:3000`.
    4.  The agent connects to the socket. MCP Any verifies the UID of the connecting process matches the server's UID.
    5.  Communication proceeds as standard MCP.

## 4. Design & Architecture
*   **System Flow:**
    - **Bootstrap**: Server checks OS and initializes either UDS or Named Pipe.
    - **Identity Verification**: Upon connection, the server uses `getsockopt` (SO_PEERCRED) on Linux or `getpeereid` on macOS to verify the peer's identity.
*   **APIs / Interfaces:**
    - New transport configuration option: `--transport=uds` or `--transport=named-pipe`.
    - SDK update to support connecting via file paths instead of URLs.
*   **Data Storage/State:** No changes to persistent state. Socket paths are transient.

## 5. Alternatives Considered
*   **Localhost with API Keys**: Still vulnerable to DNS rebinding or other browser-based attacks if not implemented perfectly.
*   **mTLS for Localhost**: High configuration overhead for local development.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a fundamental "Zero Trust" improvement for local environments. It prevents "ClawJacked"-style hijacking.
*   **Observability:** Log the UID/PID of every connecting client for auditability.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
