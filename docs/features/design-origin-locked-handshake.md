# Design Doc: Origin-Locked Local Handshake (OLH)
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
The discovery of CVE-2026-25253 (OpenClaw token exfiltration) has proven that "Implicit Local Trust" for loopback interfaces is a catastrophic failure point. Browsers can be used by malicious websites to bridge into the local control plane of AI agents via unauthenticated WebSocket or HTTP requests.

MCP Any needs a "Zero-Trust" local transport layer that mandates cryptographically bound origin validation for all local listeners. This ensures that only verified local applications (e.g., the MCP Any UI or authorized CLI) can interact with the gateway, even if the request originates from `localhost`.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate `Origin` and `Sec-Fetch-Site` header verification for all local API and WebSocket endpoints.
    * Implement a hardware-attested handshake for initial session binding.
    * Cryptographically bind session tokens to the initiating browser/process origin.
* **Non-Goals:**
    * Blocking legitimate remote access (remote access already requires full MFA/Attestation).
    * Providing general-purpose network encryption (handled by TLS/mTLS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Prevent malicious browser scripts from hijacking the agent control plane.
* **The Happy Path (Tasks):**
    1. The developer starts the MCP Any gateway.
    2. The local UI attempts to connect via WebSocket.
    3. MCP Any verifies the `Origin` header matches an allow-listed local application.
    4. A "Pairing Handshake" is initiated where the user must provide a hardware-bound approval (e.g., TouchID/TPM).
    5. A session token is issued, cryptographically pinned to that specific `Origin`.
    6. Subsequent tool calls from that UI are accepted, while calls from a random browser tab are blocked.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Browser as Browser Tab (Malicious)
        participant UI as Local Authorized UI
        participant OLH as OLH Middleware
        participant Server as MCP Any Server

        UI->>OLH: WebSocket Handshake (Origin: mcp-any://local)
        OLH->>OLH: Verify Origin & Sec-Fetch-Site
        OLH->>UI: Request Hardware Attestation
        UI->>OLH: Hardware-Bound Proof
        OLH->>Server: Establish Pinned Session

        Browser->>OLH: WebSocket Handshake (Origin: https://evil.com)
        OLH->>OLH: Validation Failed
        OLH-->>Browser: 403 Forbidden
    ```
* **APIs / Interfaces:**
    * `POST /api/v1/auth/pair`: Endpoint for initiating the hardware-locked handshake.
    * Middleware injection into the WebSocket upgrader to enforce origin-binding.
* **Data Storage/State:**
    * Session tokens are stored in-memory, mapped to origin hashes and hardware IDs.

## 5. Alternatives Considered
* **IP Whitelisting:** Rejected because all browser tabs share the same `127.0.0.1` IP, providing no protection against cross-site hijacking.
* **Mutual TLS (mTLS) for Local:** Rejected as too high-friction for standard local development, although still supported for remote enterprise nodes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enforces Origin-validation at the kernel/transport boundary before any application-level logic executes.
* **Observability:** Blocked origin violations are logged with full stylometric metadata for forensic analysis.

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
