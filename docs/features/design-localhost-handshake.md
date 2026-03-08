# Design Doc: Localhost Handshake Protocol (Anti-ClawJack)

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The "ClawJacked" exploit demonstrated that AI agents listening on local ports are vulnerable to Cross-Origin Resource Sharing (CORS) bypass and WebSocket hijacking from malicious websites running in a user's browser. Since browsers can make requests to `localhost`, a malicious site can brute-force ports and interact with an unprotected MCP Any instance. We need a way to ensure that only authorized local clients (like a CLI or a trusted UI) can communicate with the server.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Prevent unauthorized WebSocket and HTTP access to MCP Any from browser-based environments.
    *   Implement a "Handshake" mechanism that requires a client-side secret or token not accessible to arbitrary websites.
    *   Ensure minimal friction for trusted local tools (CLI, Desktop App).
*   **Non-Goals:**
    *   Securing remote connections (covered by general Zero Trust / MFA design).
    *   Replacing TLS (though TLS on localhost is a potential component).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using MCP Any via CLI and a browser-based dashboard.
*   **Primary Goal:** Access the dashboard securely without exposing the agent to malicious sites in other tabs.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any.
    2.  Server generates a short-lived `HandshakeToken` and writes it to a secure local file (e.g., `~/.mcpany/session_token`).
    3.  Trusted local dashboard reads this token and includes it in the `Sec-WebSocket-Protocol` header or an `X-MCP-Handshake` HTTP header.
    4.  Server validates the token and allows the connection.
    5.  A malicious site in another tab tries to connect but cannot read `~/.mcpany/session_token` due to OS-level file permissions and browser sandboxing, failing the handshake.

## 4. Design & Architecture
*   **System Flow:**
    - **Token Generation**: On startup, the server creates a high-entropy random token.
    - **Secure Storage**: The token is stored in a file with `0600` permissions.
    - **Verification Middleware**: A middleware intercepts all incoming requests. It checks for the handshake token.
    - **Connection Upgrade**: For WebSockets, the handshake must be completed before the protocol upgrade.
*   **APIs / Interfaces:**
    - Header: `X-MCP-Handshake: <token>`
    - WebSocket Subprotocol: `mcp-handshake.<token>`
*   **Data Storage/State:** In-memory storage of the current valid session token.

## 5. Alternatives Considered
*   **Origin Checking**: Checking the `Origin` header. *Rejected* as it can be spoofed or bypassed in some scenarios, and doesn't protect against non-browser clients.
*   **CSRF Tokens**: Standard CSRF protection. *Rejected* as it's more complex for pure API/WebSocket use cases and still relies on cookies which can be targeted.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This closes the "last mile" of localhost security.
*   **Observability:** Log failed handshake attempts with source IP and headers to identify potential "ClawJacking" attempts.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
