# Design Doc: Mandatory Origin Attestation (Localhost-Zero-Trust)
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Recent security breaches in local AI agents (e.g., OpenClaw CVE-2026-25253) have demonstrated that the `localhost` boundary is insufficient for security. Malicious websites running in a user's browser can perform cross-origin requests or DNS rebinding to hijack local agent processes that lack strict origin validation. MCP Any must ensure that only authorized clients (CLI, approved Web UIs, or verified local apps) can communicate with the gateway, even when connecting via localhost.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory cryptographic handshake for all incoming connections (WebSocket/HTTP).
    * Validate the `Origin` and `Host` headers against a strictly controlled allowlist.
    * Use short-lived, client-specific access tokens generated via a secure "pairing" flow.
    * Prevent "One-Click" hijacking from browser-based malicious scripts.
* **Non-Goals:**
    * Replacing TLS for remote connections (this is an additional layer).
    * Implementing a full identity provider (OIDC/SAML).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a local LLM and a suite of MCP tools.
* **Primary Goal:** Use MCP Any securely without worrying about a random website stealing their local credentials or executing shell commands via the agent.
* **The Happy Path (Tasks):**
    1. User starts `mcp-any` server.
    2. On first run, `mcp-any` generates a unique "Pairing Code" and saves it to a secure, OS-level keystore.
    3. User opens the MCP Any Dashboard (local UI).
    4. The Dashboard prompts the user for the Pairing Code.
    5. Upon entry, the Dashboard and Server perform a Diffie-Hellman handshake to establish a shared secret.
    6. All subsequent requests from the Dashboard include a HMAC signature of the payload using the shared secret.
    7. The Server rejects any request lacking a valid signature or coming from an unverified Browser Origin.

## 4. Design & Architecture
* **System Flow:**
    `Client (Browser/CLI) <--> [Handshake Middleware] <--> [MCP Any Core]`
    1. **Handshake**: Client sends a `HELLO` with a nonce.
    2. **Challenge**: Server responds with a challenge and its public key.
    3. **Attestation**: Client signs the challenge using the shared secret (derived from Pairing Code) and returns it.
    4. **Session**: Server issues a session-bound JWT.
* **APIs / Interfaces:**
    * `POST /auth/handshake`: Initial exchange.
    * `POST /auth/verify`: Token refresh and attestation check.
    * All MCP Tool endpoints will now require the `Authorization: Bearer <JWT>` header.
* **Data Storage/State:**
    * Pairing Codes stored in `keyring` (macOS/Linux/Windows).
    * Active sessions stored in an in-memory LRU cache with 1-hour TTL.

## 5. Alternatives Considered
* **Simple API Keys**: Rejected because they are often stored in plain text configuration files (which are susceptible to "Claude Code" style hijacking).
* **IP-based Allowlisting**: Rejected because it does not protect against browser-based attacks coming from the same machine (localhost).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements the "Never Trust, Always Verify" principle for the most vulnerable link: the local loopback.
* **Observability**: Failures in attestation will be logged with high severity, including the offending `Origin` and `User-Agent`.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
