# Design Doc: Anti-Hijack WebSocket Middleware
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
Recent security research (Oasis-2026-004) has identified a systemic vulnerability in AI agent gateways that bind to `localhost`. Browsers do not enforce Same-Origin Policy (SOP) or Cross-Origin Resource Sharing (CORS) for WebSockets. A malicious website can initiate a WebSocket connection to `localhost:[agent_port]` and execute commands if the user is logged in or if authentication is weak/absent for local connections. MCP Any must protect against this "Localhost Hijack."

## 2. Goals & Non-Goals
* **Goals:**
    * Mandatory `Origin` header validation for all WebSocket upgrade requests.
    * Explicit allow-listing of trusted origins (e.g., `vscode-webview://`, `chrome-extension://`).
    * Rate limiting for all WebSocket connection attempts and authentication requests, with zero loopback exemptions.
    * Implementation of a "Local Secret" handshake for any browser-based connection.
* **Non-Goals:**
    * Replacing standard JSON-RPC over WebSockets.
    * Handling non-WebSocket transport security (covered by other docs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a browser-based IDE extension.
* **Primary Goal:** Connect the extension to MCP Any securely without exposing the gateway to malicious tabs.
* **The Happy Path (Tasks):**
    1. User starts MCP Any.
    2. User opens their browser-based IDE (e.g., VS Code Web).
    3. The IDE extension attempts to connect to MCP Any via WebSocket.
    4. MCP Any detects the `Origin` header and checks it against the allow-list.
    5. If not allow-listed, MCP Any requires the user to copy a "One-Time Pairing Code" from the terminal to the extension.
    6. Once paired, the `Origin` + `Secret` are cached as trusted.

## 4. Design & Architecture
* **System Flow:**
    - **Upgrade Interceptor**: A new middleware in the WebSocket handler that inspects the `Upgrade` request.
    - **Origin Validator**: Compares the `Origin` header against a `trusted_origins` configuration and a dynamic `paired_origins` database.
    - **Pairing Flow**: If the origin is unknown, the connection is held in a "Pre-Auth" state, and a pairing request is logged/displayed to the user.
* **APIs / Interfaces:**
    - `GET /ws/v1/connect`: Standard WebSocket endpoint with mandatory Origin checks.
    - `POST /api/v1/auth/pair`: Endpoint to submit the One-Time Pairing Code.
* **Data Storage/State:**
    - `paired_origins.db`: SQLite table storing `(origin, public_key, trust_level)`.

## 5. Alternatives Considered
* **Disabling WebSockets entirely**: *Rejected* as many modern agent UIs and extensions rely on real-time bidirectional communication.
* **Requiring HTTPS for localhost**: *Rejected* because managing local CA certificates is a major friction point for users ("Self-signed certificate" errors).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This feature mitigates "Lateral Movement" from a compromised browser tab to the local agent infrastructure.
* **Observability:** Failed connection attempts due to Origin mismatches must be prominently logged and reported in the `doctor` command.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
