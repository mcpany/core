# Design Doc: Zero-Trust Local Handshake (ZTLH) Provider
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
The disclosure of critical loopback hijacking vulnerabilities (VU#221883) has proven that "Implicit Local Trust" is a catastrophic failure point for AI agents. Malicious websites can use browser-based scripts to bridge the gap between the public internet and local listeners (127.0.0.1), effectively hijacking the agent's control plane without user interaction.

MCP Any needs to implement the ZTLH Provider to mandate a cryptographically secure, hardware-attested handshake for all local communication. This ensures that only authorized local applications (e.g., a verified CLI or local IDE) can communicate with the gateway, neutralizing cross-site request forgery (CSRF) and WebSocket hijacking attempts.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory `Origin` and `Sec-Fetch-Site` header verification for all local listeners.
    * Use hardware-bound (TPM/SEP) keys to sign local handshake tokens.
    * Bind every agent session to a cryptographically verified initiating origin.
    * Provide a fallback "User-in-the-Loop" approval flow for un-attested local applications.
* **Non-Goals:**
    * Securing remote (non-loopback) traffic (handled by existing mTLS/FSI providers).
    * Providing full end-to-end encryption for the local bus (focus is on authentication and origin-locking).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a malicious browser tab from commanding the local MCP Any gateway.
* **The Happy Path (Tasks):**
    1. Developer starts MCP Any and a compliant local tool (e.g., Claude Code).
    2. The local tool initiates a connection to the MCP Any loopback port.
    3. ZTLH Provider intercepts the request and verifies the `Origin` header.
    4. ZTLH issues a "Handshake Challenge" requiring a hardware-attested signature from the tool.
    5. The tool signs the challenge using a local TPM-bound key.
    6. ZTLH validates the signature and establishes a session-locked WebSocket.
    7. A malicious script in a background browser tab attempts to connect; ZTLH detects the unauthorized origin/lack of signature and drops the connection.

## 4. Design & Architecture
* **System Flow:**
    `Local Tool` -> `HTTP GET /handshake (with Origin)` -> `ZTLH Provider`
    `ZTLH Provider` -> `401 Unauthorized (Nonce + Challenge)` -> `Local Tool`
    `Local Tool` -> `POST /handshake (Signed Nonce + TPM Metadata)` -> `ZTLH Provider`
    `ZTLH Provider` -> `200 OK (Origin-Bound Session Token)` -> `Local Tool`
* **APIs / Interfaces:**
    * `/v1/auth/handshake`: New endpoint for multi-factor local authentication.
    * `X-MCP-Handshake-Signature`: Header for passing hardware-attested tokens.
* **Data Storage/State:**
    * Short-lived, memory-resident "Handshake Nonces."
    * Origin-to-Session mapping in the secure session store.

## 5. Alternatives Considered
* **Static API Keys:** Rejected because they are often stored in plaintext configuration files and can be exfiltrated via local file access exploits.
* **OS-Level Permissions (e.g., unix sockets):** While secure, they are harder to implement cross-platform (Windows vs. Linux) and don't provide the same granular origin-locking as ZTLH.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZTLH implements the "Never Trust, Always Verify" model for the local intranet. It prevents "ClawJacked" style attacks by requiring both network-level (Origin) and cryptographic (TPM) proof of identity.
* **Observability:** All failed handshake attempts are logged with origin metadata and flagged in the `Local Security Violation Monitor`.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
