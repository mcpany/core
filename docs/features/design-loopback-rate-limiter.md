# Design Doc: Local-Loopback Rate Limiter
**Status:** Draft
**Created:** 2026-03-17

## 1. Context and Scope
Today's market sync revealed a critical vulnerability (CVE-2026-25253) where malicious browser-based content can bridge to local AI agent listeners. Furthermore, the Oasis Security report confirmed that most local agent infrastructures treat `127.0.0.1` as a trusted zone, exempting it from rate limiting. This allows malicious JavaScript to brute-force gateway credentials at machine speeds.

The **Local-Loopback Rate Limiter** is a mandatory security middleware for MCP Any that enforces strict request throttling and origin-based auditing for all traffic arriving on the loopback interface.

## 2. Goals & Non-Goals
*   **Goals:**
    - Enforce per-origin rate limits on all loopback (127.0.0.1 / ::1) API and WebSocket endpoints.
    - Implement mandatory audit logging for all blocked loopback requests.
    - Provide a "Zero-Trust" posture for local development without significantly impacting developer ergonomics.
*   **Non-Goals:**
    - This system will NOT manage authentication; it only handles request frequency and origin validation.
    - It will NOT intercept low-level kernel-to-kernel communication that doesn't pass through the network stack.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local Developer / End User
*   **Primary Goal:** Prevent a malicious website from brute-forcing the local MCP Any password or session token via hidden fetch/WebSocket calls.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any server with default "Local Zero Trust" settings.
    2.  A malicious site attempts 100 WebSocket connections per second to `localhost:3000`.
    3.  The Rate Limiter identifies the burst and the untrusted `Origin`.
    4.  Subsequent requests from that origin are immediately dropped with a `429 Too Many Requests`.
    5.  The incident is logged in the `Local Security Audit Dashboard`.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Client[Browser / CLI Client] -->|Request| Listener[MCP Any Listener]
        Listener -->|IP/Origin| RL[Loopback Rate Limiter Middleware]
        RL -->|Check Quota| Store[In-Memory Token Bucket Store]
        RL -->|If Exceeded| Block[429 Too Many Requests / Close WS]
        RL -->|If Valid| App[Core MCP Server Logic]
        Block --> Audit[Security Audit Log]
    ```
*   **APIs / Interfaces:**
    - `RateLimitConfig`: YAML schema for defining `requests_per_second` and `burst_size` specifically for `127.0.0.1`.
    - `AuditLogger`: Internal interface for streaming blocked request metadata to the UI.
*   **Data Storage/State:**
    - Uses an in-memory, thread-safe `map[string]*RateLimiter` where the key is the `Origin` header (or IP if missing).

## 5. Alternatives Considered
*   **IP-only Throttling**: Rejected because multiple browser origins can share the same local IP (127.0.0.1), making it impossible to distinguish between a legitimate local UI and a malicious external site.
*   **OS-level Firewall (iptables)**: Rejected because it cannot inspect HTTP `Origin` headers required to distinguish browser-based attacks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The limiter acts as the first line of defense in a Local Zero Trust model.
*   **Observability:** All throttled events must be exposed via the management API for real-time visualization in the `Local Security Violation Monitor`.

## 7. Evolutionary Changelog
*   **2026-03-17:** Initial Document Creation.
