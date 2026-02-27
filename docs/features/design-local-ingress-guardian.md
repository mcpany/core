# Design Doc: Local Ingress Guardian

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent critical vulnerabilities in OpenClaw (and other local-first agent gateways) have demonstrated that binding a WebSocket or HTTP server to `localhost` is insufficient for security. Malicious websites can initiate WebSocket connections to `localhost` from a user's browser, bypassing standard password protections and gaining full control over the agent. MCP Any must provide a more secure method for local agents to communicate without exposure to browser-origin attacks.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a communication channel for local LLMs/Agents that is inaccessible to web browsers.
    *   Support Unix Domain Sockets (UDS) on Linux/macOS and Named Pipes on Windows.
    *   Implement cryptographically bound mTLS for local network communication where IPC is not possible.
    *   Integrate with the Policy Firewall to ensure only authorized local processes can connect.
*   **Non-Goals:**
    *   Securing remote (cloud-to-local) connections (handled by Environment Bridging Middleware).
    *   Replacing HTTP/WS entirely (legacy support remains).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local Developer / Privacy-Conscious User.
*   **Primary Goal:** Use a local coding agent (e.g., OpenClaw or Claude Code) without the risk of a malicious website hijacking the agent session.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any with the `--secure-ingress` flag.
    2.  MCP Any creates a Unix Domain Socket at `~/.mcpany/ingress.sock` with restricted permissions (0600).
    3.  The User configures their agent (e.g., Cursor or a custom script) to connect via the socket rather than `http://localhost:8080`.
    4.  Browser-based JavaScript attempts to connect to the socket but fails (browsers cannot access the filesystem/UDS).
    5.  The local agent communicates securely and with lower latency.

## 4. Design & Architecture
*   **System Flow:**
    - **Bootstrap**: At startup, the `IngressManager` determines the OS and initializes the appropriate IPC mechanism.
    - **Authentication**: Connection is implicitly authenticated by filesystem permissions (UID/GID check). Optional HMAC-based handshake for additional security.
    - **Protocol Bridge**: Incoming IPC traffic is decoded and routed into the standard MCP Any middleware pipeline.
*   **APIs / Interfaces:**
    - `InternalIPCListener`: Interface for UDS/Named Pipes.
    - `SecureTransportWrapper`: Handles optional encryption/auth on top of the transport.
*   **Data Storage/State:** None required (stateless transport layer).

## 5. Alternatives Considered
*   **Strict CORS/Host Checking**: *Rejected* because WebSockets do not strictly follow CORS, and Host header spoofing/rebinding is a known bypass vector for many implementations.
*   **Custom Browser Extensions**: *Rejected* because it increases friction and doesn't protect users who don't install the extension.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature is a foundational "Zero Trust" component for local execution. It eliminates an entire class of remote-to-local attacks.
*   **Observability:** Log connection attempts from unauthorized UIDs or malformed IPC packets.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
