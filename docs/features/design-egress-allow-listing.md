# Design Doc: Context-Aware Egress Allow-Listing
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
With the discovery of CVE-2026-21852 (Base URL Hijacking), agents can be coerced into exfiltrating sensitive API keys by redirecting traffic to malicious domains. MCP Any must act as a mandatory egress gateway that validates every outbound request against the agent's current "Intent-Scope." This ensures that even if an agent's configuration is compromised, it cannot send data to unauthorized endpoints.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all agent outbound traffic (LLM providers, MCP servers).
    * Enforce a strict allow-list of domains based on the active "Intent-Scope."
    * Provide real-time alerts for blocked exfiltration attempts.
* **Non-Goals:**
    * Inspecting encrypted payload content (DLP is a separate concern).
    * Managing the LLM provider's internal routing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent API key exfiltration from autonomous agents.
* **The Happy Path (Tasks):**
    1. Architect defines an "Intent-Scope" for a project (e.g., "Web Research").
    2. MCP Any generates an allow-list for that scope (e.g., `google.com`, `wikipedia.org`, `api.anthropic.com`).
    3. Agent attempts to call a malicious domain `evil.com`.
    4. MCP Any Interceptor detects the mismatch with "Web Research" scope and severs the connection.
    5. Security Architect receives a high-priority alert in the dashboard.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [MCP Any Proxy] -> [Egress Filter] -> [Target Domain]`
    The Egress Filter checks the target domain against a `ScopeRegistry`.
* **APIs / Interfaces:**
    * `GET /v1/security/egress/policies`: List active egress rules.
    * `POST /v1/security/egress/verify`: Check if a domain is allowed for a given token.
* **Data Storage/State:**
    Policies are stored in the Service Registry and cached in-memory for low-latency filtering.

## 5. Alternatives Considered
* **Host-level Firewall (iptables):** Rejected because it lacks "Agent" or "Intent" context.
* **DNS Filtering:** Rejected because it can be bypassed via hardcoded IP addresses.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses short-lived capability tokens to bind requests to specific scopes.
* **Observability:** All blocked requests are logged to the `Audit Log` with full context.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
