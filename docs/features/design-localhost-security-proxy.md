# Design Doc: Zero-Trust Localhost Guard (ClawJacked Mitigation)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The "ClawJacked" exploit (CVE-2026-25253) demonstrated that the traditional "localhost is safe" assumption in developer tools is fundamentally flawed. Modern browsers allow external websites to initiate WebSocket and HTTP connections to local services. If these local services (like MCP Any or OpenClaw) disable security for `localhost` or `127.0.0.1`, a malicious website can hijack the agent, gaining full access to the user's system and credentials.

MCP Any must implement a hardening layer that treats localhost as an untrusted network, requiring explicit, cryptographically-verified pairing for all incoming connections.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce mandatory `Origin` header verification for all incoming HTTP/WebSocket requests.
    * Implement a "Device Pairing" flow requiring a one-time cryptographic handshake for new clients.
    * Support "Silent Re-pairing" for previously authorized local clients via secure, short-lived tokens.
    * Provide a clear UI for users to manage authorized local clients.
* **Non-Goals:**
    * Providing a full identity provider (IdP) for local users.
    * Encrypting localhost traffic (MTLS) unless specifically requested (performance trade-off).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer using OpenClaw + MCP Any.
* **Primary Goal:** Securely connect a local agent to MCP Any without exposing the gateway to malicious browser tabs.
* **The Happy Path (Tasks):**
    1. User starts MCP Any.
    2. User attempts to connect OpenClaw to MCP Any via WebSocket.
    3. MCP Any detects an unauthorized connection attempt.
    4. MCP Any UI displays a "Pairing Request" with a 6-digit PIN.
    5. User enters the PIN into the OpenClaw configuration or approves via a CLI prompt.
    6. MCP Any issues a `Client-Token` that is stored in the client's local configuration.
    7. Subsequent connections use the `Client-Token` for silent authorization.

## 4. Design & Architecture
* **System Flow:**
    1. **Listener**: Intercepts all incoming connections.
    2. **Origin Check**: Rejects any request with an unexpected or missing `Origin` header (unless configured for specific non-browser clients).
    3. **Auth Check**: Checks for `X-MCP-Any-Token` header or `token` query param.
    4. **Pairing Service**: If no token, triggers a pairing flow (SSE or Long-Poll to notify UI/CLI).
    5. **Token Issue**: Upon PIN verification, generates a SHA-256 HMAC token bound to the client's IP and User-Agent.
* **APIs / Interfaces:**
    * `GET /v1/auth/pair`: Initiates pairing, returns `request_id`.
    * `POST /v1/auth/verify`: Accepts `request_id` and `pin`, returns `client_token`.
    * `WS /v1/mcp?token=[token]`: Secure WebSocket endpoint.
* **Data Storage/State:** Authorized client tokens are stored in a local, encrypted SQLite table (`authorized_clients`).

## 5. Alternatives Considered
* **Strict IP Whitelisting**: Rejected because "ClawJacked" specifically uses the browser to act as the "localhost" proxy, making IP whitelisting ineffective.
* **Mutual TLS (mTLS)**: Considered but rejected as the default due to the complexity of certificate management for casual local developers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses a "deny-by-default" posture for all connections. Tokens are rotation-capable and session-bound.
* **Observability:** Every pairing attempt and failed authorization is logged to the Audit Log.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
