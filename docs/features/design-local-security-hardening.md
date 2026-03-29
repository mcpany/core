# Design Doc: Local Security Hardening (Zero-Trust Loopback)

**Status:** Draft
**Created:** 2026-03-17

## 1. Context and Scope
The March 2026 Oasis Security report revealed that OpenClaw's implicit trust of local loopback connections (`127.0.0.1`) created a "Super-Highway" for attackers. By exempting localhost from rate limiting and logging, the gateway allowed malicious websites to brute-force passwords silently. MCP Any must harden its local transport layer to treat all loopback traffic with the same level of scrutiny as remote traffic.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement mandatory rate limiting for all local listeners (HTTP and WebSockets).
    *   Enforce `Origin` and `Sec-Fetch-Site` validation for all local requests.
    *   Log all local connection attempts (success and failure) with origin metadata.
    *   Provide a "Local Security Dashboard" for users to review blocked attempts.

![Dashboard](../screenshots/dashboard_overview.png)
*   **Non-Goals:**
    *   Blocking legitimate local applications (e.g., a local IDE).
    *   Implementing OS-level firewall rules.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local Developer using an MCP-capable IDE.
*   **Primary Goal:** Connect the IDE to MCP Any securely without exposing the gateway to malicious websites.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any.
    2.  IDE connects to `localhost:8080`. MCP Any validates the `Origin` header (e.g., `vscode-webview://...`).
    3.  A malicious website attempts to brute-force the password via a hidden WebSocket.
    4.  MCP Any's `LocalRateLimiter` detects the high-frequency attempts and throttles the IP `127.0.0.1`.
    5.  The attempts are logged and displayed in the UI as "Blocked Origin: [attacker.com]".

## 4. Design & Architecture
*   **System Flow:**
    - **Middleware Chain**: Every request passes through `OriginValidator` -> `RateLimiter` -> `Authenticator`.
    - **OriginValidator**: Rejects any request with a missing or non-allowlisted `Origin` header.
    - **LocalRateLimiter**: Specifically tracks `127.0.0.1` and `::1` with a lower threshold for unauthenticated requests.
*   **APIs / Interfaces:**
    - `GET /api/v1/security/local-audit`: Returns a list of recent local connection attempts.
*   **Data Storage/State:**
    - Rate limit counters stored in in-memory Radix tree.
    - Security logs stored in the local SQLite database.

## 5. Alternatives Considered
*   **Binding to Unix Domain Pipes Only**: *Rejected* because it breaks compatibility with browser-based tools and many existing IDE plugins.
*   **Implicit Trust for PIDs**: *Rejected* because mapping PIDs to network connections is complex and platform-dependent.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This design eliminates the "Local Loophole."
*   **Performance**: Rate limiting must be high-performance to avoid adding latency to legitimate local tool calls.

## 7. Evolutionary Changelog
*   **2026-03-17:** Initial Document Creation.
*   **2026-03-31:** Update: Mitigating CVE-2026-34812 (Deep Symlink Escape).
    *   **Context:** Today's market sync revealed a new exploit where recursive symlinks in project-local configs allow sandbox escape.
    *   **Architecture Adjustment:** Introducing the `Inode-Aware Symlink Validator` into the configuration discovery pipeline.
    *   **Security Impact:** Prevents filesystem traversal outside of the project root by resolving all symlinks and validating target inodes against an "Approved Root" registry before any file I/O or tool discovery occurs.
