# Design Doc: Local-Only WebSocket Auth (LOWA) Gateway
**Status:** Draft
**Created:** 2026-05-22

## 1. Context and Scope
The "ClawJacked" vulnerability (CVE-2026-25253) demonstrated that the assumption of "Implicit Local Trust" for loopback traffic is fundamentally flawed. Malicious websites can use JavaScript to establish WebSocket connections to `localhost`, bypassing same-origin policies for standard HTTP requests. If the local gateway does not enforce strong, rate-limited authentication for these connections, an attacker can hijack the entire agent control plane.

LOWA is a mandatory security layer for MCP Any that enforces session-bound authentication for all local WebSocket listeners. It ensures that only authorized local applications—not rogue scripts in a browser tab—can command the Universal Agent Bus.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Mandate cryptographic authentication for all loopback (`127.0.0.1`, `::1`) WebSocket connections.
    *   Implement strict rate limiting for authentication attempts on local ports.
    *   Bind session tokens to specific browser/CLI origins.
    *   Provide a "One-Click Approve" flow for initial local pairing that requires manual user confirmation.
*   **Non-Goals:**
    *   Replacing standard mTLS for remote connections.
    *   Providing a full-blown identity provider (LOWA is for local session bootstrapping).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using MCP Any with a local browser-based IDE (e.g., VS Code Web).
*   **Primary Goal:** Securely connect the local IDE to the MCP Any gateway without exposing it to malicious sites.
*   **The Happy Path (Tasks):**
    1.  The user starts MCP Any.
    2.  The IDE attempts to connect via WebSocket.
    3.  MCP Any intercepts the connection and requires an `Authorization` header or a specific handshake token.
    4.  Since it's a new connection, MCP Any triggers a system notification/terminal prompt asking for approval.
    5.  The user approves the pairing.
    6.  MCP Any issues a session-bound token to the IDE.
    7.  Subsequent messages are authorized via this token.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Browser as Browser/IDE
        participant LOWA as LOWA Gateway
        participant Core as MCP Any Core
        Browser->>LOWA: WebSocket Connect (localhost:port)
        LOWA-->>Browser: 401 Unauthorized (Challenge)
        Browser->>LOWA: Auth Request (Pairing Code/Token)
        LOWA->>User: Desktop/CLI Prompt: "Approve connection from X?"
        User-->>LOWA: Approve
        LOWA->>Browser: Session Token (Origin-Bound)
        Browser->>LOWA: Command (with Token)
        LOWA->>Core: Process Command
    ```
*   **APIs / Interfaces:**
    *   `GET /ws`: Upgrades to WebSocket, requires `Sec-WebSocket-Protocol: mcp-any-v1-auth` or a token query param.
    *   `POST /auth/pair`: Initiates the pairing flow for new local clients.
*   **Data Storage/State:**
    *   Ephemeral store for active session tokens, mapping `Token -> {Origin, Expiry, Permissions}`.

## 5. Alternatives Considered
*   **Standard Origin Header Checking:** Rejected because `Origin` can sometimes be spoofed or bypassed in certain WebSocket scenarios, and it doesn't protect against brute-force password guessing.
*   **Static Password:** Rejected as the primary mechanism because it's vulnerable to the brute-force chain described in "ClawJacked." LOWA adds a required user-in-the-loop pairing step.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** LOWA is a core component of the "Local Zero Trust" mandate. It treats loopback as untrusted until proven otherwise.
*   **Observability:** All failed local auth attempts are logged to the `Local Security Audit Dashboard`.

## 7. Evolutionary Changelog
*   **2026-05-22:** Initial Document Creation.
