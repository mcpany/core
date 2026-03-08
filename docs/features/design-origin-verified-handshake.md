# Design Doc: Origin-Verified Gateway Handshake
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
The "ClawJacked" vulnerability demonstrated that local WebSocket servers (typical for MCP gateways) can be hijacked by malicious websites via a user's browser. Since browsers do not block cross-origin WebSocket connections to `localhost`, a site can brute-force or use default credentials to control a local AI agent. MCP Any must implement a mechanism to verify that a connection request originates from a trusted client application, not a browser-based script.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent 0-click WebSocket hijacking from browsers.
    * Implement a cryptographic challenge-response for all local gateway connections.
    * Support "App Binding" where specific trusted apps (e.g., Claude Code, OpenClaw Desktop) can pre-register.
* **Non-Goals:**
    * Implementing a full OIDC/OAuth2 server.
    * Replacing existing API key authentication (this is an additional layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using Claude Code with MCP Any.
* **Primary Goal:** Use local tools securely without worrying about malicious websites stealing API keys or executing commands.
* **The Happy Path (Tasks):**
    1. User starts MCP Any gateway.
    2. User starts Claude Code.
    3. Claude Code initiates a handshake with MCP Any.
    4. MCP Any sends a unique, short-lived challenge.
    5. Claude Code signs the challenge using a local private key (stored in the OS keychain) and returns it.
    6. MCP Any verifies the signature against a pre-registered public key or a trusted "App Registry".
    7. Connection is upgraded to WebSocket only after successful verification.

## 4. Design & Architecture
* **System Flow:**
    1. `GET /connect` (HTTP) with `X-App-ID` header.
    2. Gateway responds with `401 Unauthorized` + `X-Challenge: [nonce]`.
    3. Client sends `GET /connect` with `X-Response: [signed-nonce]`.
    4. Gateway verifies and upgrades to `101 Switching Protocols` (WebSocket).
* **APIs / Interfaces:**
    * New middleware: `OriginVerifierMiddleware`.
    * Challenge generation service using secure random nonces.
* **Data Storage/State:**
    * Short-lived in-memory store for pending challenges (TTL < 5s).
    * `trusted_apps.json` or SQLite table for storing public keys of authorized local applications.

## 5. Alternatives Considered
* **CORS/Origin Checking:** Rejected because malicious sites can spoof simple headers, and WebSocket `Origin` headers are helpful but not sufficient against all bypasses.
* **Loopback-Only Binding:** Already implemented but doesn't solve the browser-to-localhost attack vector.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Even if the password/token is compromised, the attacker still needs the local app's private key to establish a connection.
* **Observability:** Log failed handshake attempts with source IP and reported App-ID to detect scanning/brute-force.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
