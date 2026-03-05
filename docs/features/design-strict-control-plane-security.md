# Design Doc: Strict Control Plane Security (Anti-ClawJack)

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The "ClawJacked" exploit (CVE-2026-25253) demonstrated that AI agent control planes—specifically those using WebSockets for real-time communication—are vulnerable to standard web attacks like Cross-Site Request Forgery (CSRF). When an agent's local management interface is exposed or lacks proper origin validation, a malicious website can trigger tool executions on the user's machine. MCP Any must implement industrial-grade security to protect the agentic control plane.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce `Strict-Origin` validation for all WebSocket handshakes.
    *   Implement mandatory CSRF tokens for all state-changing HTTP and WebSocket operations.
    *   Support "Secure Context" only (HTTPS/WSS) for any remote-authorized listeners.
    *   Provide a "One-Time-Token" (OTT) mechanism for CLI-to-GUI handoffs.
*   **Non-Goals:**
    *   Implementing a full OAuth2 provider (we will use simpler, token-based attestation).
    *   Securing the tools themselves (that is the job of the Policy Firewall and DTI).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local AI Developer using a web-based dashboard to manage local agents.
*   **Primary Goal:** Use the dashboard securely without risk of CSRF attacks from other browser tabs.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any via CLI.
    2.  CLI generates a unique, short-lived `Management-Token`.
    3.  User opens the MCP Any UI; the UI must include this token in its initial handshake.
    4.  Every subsequent WebSocket message must include a sequence-aware CSRF nonce.
    5.  A malicious tab attempts to call `ws://localhost:3000/api/tools/call`; MCP Any rejects it due to missing token and invalid Origin.

## 4. Design & Architecture
*   **System Flow:**
    - **Handshake Guard**: The `WSServer` intercepts the `Upgrade` request. It validates the `Origin` header against a whitelist (defaulting to `localhost` and the configured UI domain).
    - **Token Validation**: Requests without a valid `X-MCP-Management-Token` or OTT query parameter are rejected with 403 Forbidden.
    - **Nonce Tracking**: For stateful WebSocket sessions, a rolling nonce is used to prevent replay attacks and cross-tab interference.
*   **APIs / Interfaces:**
    - `GET /api/auth/token`: CLI-only endpoint to retrieve a new management token.
    - `X-CSRF-Token`: Mandatory header for all POST/PUT/DELETE requests.
*   **Data Storage/State:** Tokens are stored in-memory (short-lived) or in the secure instance state file.

## 5. Alternatives Considered
*   **Basic Auth**: Using username/password. *Rejected* because it's prone to credential stuffing and less convenient for automated CLI-to-GUI flows.
*   **IP Whitelisting**: Restricting access to `127.0.0.1`. *Rejected* as it doesn't protect against CSRF from the same machine's browser.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a "Defense in Depth" measure. Even if the listener is exposed (violating the Safe-by-Default policy), the Control Plane remains secure.
*   **Observability:** Log all blocked handshake attempts with details on failed Origin or missing tokens for security auditing.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
