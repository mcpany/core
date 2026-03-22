# Design Doc: Local Sovereignty Gate (LSG)
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) revealed that "Implicit Local Trust" for loopback listeners is a catastrophic failure point. Malicious websites can bridge into local agent control planes via unauthenticated WebSocket connections to `localhost`. MCP Any needs a unified, non-bypassable security gate that enforces session-bound authentication and strict origin validation for all local endpoints to prevent browser-to-local hijacking.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate session-bound authentication for all local API and WebSocket listeners.
    * Enforce strict `Origin` and `Sec-Fetch-Site` header validation.
    * Neutralize cross-site hijacking and brute-force attempts on local gateways.
    * Provide a standardized "Local Sovereignty" posture for all connected adapters.
* **Non-Goals:**
    * Does not replace remote mTLS/TLS requirements for cloud-to-local bridging.
    * Does not handle tool-level permission gating (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely connect a local web UI to the MCP Any gateway without exposing it to malicious scripts on other browser tabs.
* **The Happy Path (Tasks):**
    1. User starts MCP Any with LSG enabled.
    2. MCP Any generates a session-bound "Sovereignty Token."
    3. The authorized local UI includes the token in the connection handshake.
    4. LSG verifies the token and the `Origin` header.
    5. Connection is established; unauthorized origins are immediately blocked.

## 4. Design & Architecture
* **System Flow:**
    `Local Application` -> `[Origin/Token Check (LSG)]` -> `MCP Any Gateway`
* **APIs / Interfaces:**
    * Middleware interceptor for all HTTP/WebSocket handlers.
    * Internal `SovereigntyManager` for token lifecycle and origin allow-listing.
* **Data Storage/State:**
    * Ephemeral, in-memory store for session-bound tokens.
    * Local configuration for authorized origin patterns.

## 5. Alternatives Considered
* **Relying on OS-level permissions:** Rejected because it doesn't prevent browser-based cross-site attacks.
* **Using random ports:** Rejected; security through obscurity is insufficient against port-scanning scripts.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** LSG is a core component of the Zero Trust Local Transport mandate.
* **Observability:** All failed origin/token checks are logged with high priority for audit and alerting.

## 7. Evolutionary Changelog
* **2026-03-22:** Initial Document Creation.
