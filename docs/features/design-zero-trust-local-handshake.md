# Design Doc: Zero-Trust Local Handshake Proxy
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The disclosure of CVE-2026-25253 (OpenClaw) and CVE-2026-0628 (Gemini Live) has proven that "Implicit Local Trust" is a catastrophic vulnerability. Malicious browser extensions and scripts can bridge the loopback interface to hijack agent control planes or exfiltrate sensitive local data.

The **Zero-Trust Local Handshake Proxy** is a mandatory security layer for MCP Any that enforces origin-locked, hardware-attested handshakes for all local WebSocket and API listeners. It ensures that only verified local applications—not arbitrary browser scripts—can command the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory `Origin` and `Sec-Fetch-Site` header verification for all loopback traffic.
    * Provide a session-bound cryptographic handshake for local applications.
    * Integration with TPM/Secure Enclave to issue "Local Authority" tokens.
    * Neutralize cross-site brute-force and token exfiltration attacks.
* **Non-Goals:**
    * Providing general-purpose network firewalling (handled by OS/Docker).
    * Managing remote (non-local) authentication (handled by A2A Messaging Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer using Claude Code with MCP Any.
* **Primary Goal:** Securely connect the local agent to the MCP Any gateway without exposing the control plane to malicious websites.
* **The Happy Path (Tasks):**
    1. The developer launches MCP Any; the Handshake Proxy starts on `127.0.0.1`.
    2. Claude Code attempts to connect; it first performs a `POST /auth/handshake` providing its application identity.
    3. The Proxy validates the request origin and prompts the user (via OS notification or UI) to approve the "Local Pairing."
    4. Upon approval, the Proxy issues a TPM-signed, session-bound Handshake Token.
    5. Claude Code uses this token in the `Authorization` header for all subsequent WebSocket/API calls.
    6. A malicious script in an open browser tab attempts to connect; the Proxy blocks it immediately due to missing/invalid origin and token.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant App as Local App (Claude Code)
        participant Proxy as Handshake Proxy
        participant User as User (Approval)
        participant TPM as Hardware Enclave

        App->>Proxy: POST /auth/handshake (Origin: localhost)
        Proxy->>User: Display Pairing Code
        User-->>Proxy: Approve Pairing
        Proxy->>TPM: Sign Session Token
        TPM-->>Proxy: Session-Bound Token
        Proxy-->>App: 200 OK (Handshake Token)

        App->>Proxy: WS /v1/tools (Auth: <Token>)
        Proxy-->>App: Connection Established
    ```
* **APIs / Interfaces:**
    * `POST /auth/handshake`: Initiate pairing and retrieve session token.
    * `GET /auth/status`: Check current origin-validation status.
* **Data Storage/State:**
    * **Active Pairings:** In-memory store of hardware-attested session tokens and their bound origins.

## 5. Alternatives Considered
* **Static API Keys:** Rejected as they are prone to exfiltration via `.env` files or local log snooping.
* **OS-Level Firewalls:** Complements this doc but cannot perform application-level origin validation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pairing must be user-interactive to prevent automated "Phantom Pairing" by malware.
* **Observability:** Integrated with the "Local Trust Violation Monitor" for real-time alerting on blocked hijacking attempts.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
