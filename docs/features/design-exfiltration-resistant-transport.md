# Design Doc: Exfiltration-Resistant Transport
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
Standard AI agents often communicate directly with LLM providers (Anthropic, OpenAI) via HTTP. However, recent vulnerabilities (CVE-2026-21852) show that attackers can redirect this traffic to malicious endpoints by modifying local configuration files, leading to API key theft. MCP Any needs to enforce a "Locked Transport" where all agent outbound traffic is proxied through a secure, allow-listed gateway that prevents redirection to unauthorized domains.

## 2. Goals & Non-Goals
* **Goals:**
    * Force all agent LLM and Tool traffic through the MCP Any proxy.
    * Maintain a strict "Allow-List" of authorized upstream domains (e.g., `*.anthropic.com`, `*.openai.com`).
    * Transparently intercept and validate API base URL configurations.
    * Provide real-time alerts for any attempted exfiltration or "Base URL Hijacking."
* **Non-Goals:**
    * Deep inspection of LLM payloads (focus is on routing and transport security).
    * Providing a general-purpose VPN or system-wide proxy.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer working in a potentially compromised environment.
* **Primary Goal:** Ensure API keys are never sent to a non-official endpoint, even if the agent configuration is maliciously altered.
* **The Happy Path (Tasks):**
    1. Agent starts and attempts to load its default configuration.
    2. MCP Any intercepts the config and rewrites `ANTHROPIC_BASE_URL` to `http://localhost:MCP_ANY_PROXY`.
    3. The agent sends an API request to the proxy.
    4. MCP Any validates that the intended destination (set in the proxy's internal routing table) matches the allow-list.
    5. The request is securely forwarded to the official Anthropic API.

## 4. Design & Architecture
* **System Flow:**
    `Agent (Runtime)` -> `Rewritten Config (Localhost Proxy)` -> `MCP Any Transport Gateway` -> `Authorized Upstream (TLS)`
    1. **Proxy Injection**: MCP Any identifies the agent's transport configuration and redirects it to itself.
    2. **Egress Validation**: The Gateway checks the `Host` header and destination IP against a verified allow-list.
    3. **Credential Wrapping**: Sensitive headers (API keys) are only attached by the Gateway *after* the destination is verified, preventing the agent from ever sending keys directly to an untrusted endpoint.
* **APIs / Interfaces:**
    * `HTTP Proxy (Connect/Forward)`: Standard proxy interface for agent runtimes.
    * `Control Plane API`: To manage and attest the allow-list of upstream domains.
* **Data Storage/State:**
    * `allow_list.yaml`: Configuration-driven list of trusted endpoints.
    * `Egress Logs`: Audit trail of all outbound traffic from agents.

## 5. Alternatives Considered
* **mTLS for all connections**: Ideal but difficult to implement as it requires the LLM providers to support client certificates from individual developer machines.
* **OS-level Firewall (e.g., Little Snitch style)**: Too intrusive and difficult to configure automatically for every developer project.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Credential wrapping ensures that even if an agent is "pwned" at the runtime level, it cannot leak the master API key because it only ever sees the "Session Token" provided by MCP Any.
* **Observability**: The UI will provide a "Security Traffic Map" showing all outbound agent requests and their attestation status.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
