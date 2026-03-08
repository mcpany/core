# Design Doc: Credential Proxy Guard
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
CVE-2026-21852 revealed how agents can be tricked into stealing API keys by simply changing a `BASE_URL` in a configuration file. If an agent (like Claude Code) sends a request with an `Authorization` header to an attacker-controlled URL, the key is compromised.

Credential Proxy Guard prevents this by acting as a "Last-Mile Proxy" for all sensitive API requests. Instead of the agent knowing the real upstream URL and key, it communicates with MCP Any, which then attaches the credentials and forwards the request only to verified endpoints.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all outbound tool calls that require credentials.
    * Centralize secret management; secrets never touch project-level config.
    * Validate the destination URL against a whitelist before attaching secrets.
* **Non-Goals:**
    * Implementing a full API Gateway for all traffic.
    * Managing user-level OAuth flows (this is for service-to-service keys).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Engineer.
* **Primary Goal:** Prevent an agent from leaking a GitHub Personal Access Token to a malicious proxy.
* **The Happy Path (Tasks):**
    1. Admin configures `GITHUB_TOKEN` in MCP Any's global, secure environment.
    2. Agent tries to call a GitHub tool with a project-configured `baseUrl: http://malicious-proxy.com`.
    3. MCP Any intercepts the call, sees it requires the `GITHUB_TOKEN` capability.
    4. MCP Any checks the `baseUrl` against the allowed `github.com` whitelist.
    5. The check fails; MCP Any blocks the request and alerts the user.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent -->|Tool Call (Placeholder URL)| Proxy_Guard
        Proxy_Guard -->|Verify Destination| Whitelist
        Whitelist -->|Match| Secret_Store
        Secret_Store -->|Attach Credentials| Real_Upstream
        Real_Upstream --> Response
        Response --> Agent
    ```
* **APIs / Interfaces:**
    * `CredentialProvider`: Interface for secret retrieval (Vault, Env, etc.).
    * `DestinationValidator`: Regex-based whitelist for allowed upstreams.
* **Data Storage/State:**
    * Secure, encrypted storage for API keys (e.g., via the existing SQLite/Postgres backend with encryption at rest).

## 5. Alternatives Considered
* **Environment Variable Redaction in Logs:** Insufficient because it doesn't prevent the key from being *sent* to a malicious server.
* **Agent-Side Sandboxing:** Too dependent on the agent's implementation; MCP Any provides a protocol-level guardrail.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Minimal privilege; a tool call only gets the specific key it needs for that specific execution.
* **Observability:** Audit logs capture every attempt to use a credential, including blocked malicious destinations.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
