# Design Doc: Loopback Authentication Proxy

**Status:** Draft
**Created:** 2026-05-13

## 1. Context and Scope
The "ClawdBot" unauthenticated loopback vulnerability (Guardz 2026) revealed that legacy agents often assume anything connecting from `127.0.0.1` is trusted. This allows any local process (including malicious scripts or browser-based attacks) to command an agent. The Loopback Authentication Proxy is a mandatory transitionary layer that enforces origin-locked authentication for all remaining local network ports during the migration to "Port-Free" transport.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all incoming traffic on legacy local ports (e.g., 18789).
    * Enforce mandatory authentication for all loopback connections.
    * Validate `Origin` and `Sec-Fetch-Site` headers for all browser-initiated requests.
    * Log and block unauthorized local connection attempts.
* **Non-Goals:**
    * Providing a permanent solution for local ports (they are still being deprecated).
    * Securing remote (non-loopback) traffic (handled by existing mTLS/Auth layers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer
* **Primary Goal:** Use a browser-based dashboard to view agent status without allowing malicious websites to command the agent.
* **The Happy Path (Tasks):**
    1. The user opens the Control UI on `localhost:18789`.
    2. The browser sends a request with an `Origin` header.
    3. The Loopback Authentication Proxy intercepts the request.
    4. The Proxy verifies that the `Origin` matches an allow-listed local app.
    5. The Proxy requires a session-bound token before forwarding the request to the gateway.

## 4. Design & Architecture
* **System Flow:**
    * The Proxy sits in front of the existing WebSocket and HTTP servers.
    * It acts as a gatekeeper, terminating the connection if the authentication or origin check fails.
* **APIs / Interfaces:**
    * `RegisterOrigin(origin_url)`: Adds an authorized local application to the allow-list.
    * `VerifySession(token, origin)`: Validates the session token against the request origin.
* **Data Storage/State:** Session tokens are stored in memory and bound to the client origin.

## 5. Alternatives Considered
* **Disabling Loopback Entirely**: Rejected because many existing developer tools and dashboards rely on it for UI interaction.
* **Kernel-Level Port Blocking**: Rejected as too intrusive for a general-purpose agent gateway.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Mandatory origin-binding prevents cross-site hijacking even within the local network.
* **Observability**: Blocked origin attempts and unauthenticated loopback hits are visualized in the "Origin Violation Security Hub."

## 7. Evolutionary Changelog
* **2026-05-13:** Initial Document Creation.
