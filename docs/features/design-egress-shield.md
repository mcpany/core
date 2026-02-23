# Design Doc: Zero-Trust Egress Shield
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
Recent critical vulnerabilities in agent frameworks like OpenClaw (specifically CVE-2026-26322) have demonstrated that AI agents can be manipulated into performing Server-Side Request Forgery (SSRF) attacks. By passing malicious URLs as tool parameters, an attacker can use the agent's identity to probe internal networks, access cloud metadata services (e.g., IMDSv2), or exfiltrate data. MCP Any, as a universal gateway, must provide a "Shield" that intercepts and validates all outbound requests triggered by tool execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory egress filtering layer for all HTTP-based tool calls.
    * Provide a default-deny policy for private IP ranges (RFC 1918) and cloud metadata IPs.
    * Support domain-based allowlisting and denylisting via configuration.
    * Integrate with the `Policy Firewall` (Rego/CEL) for dynamic egress decisions.
* **Non-Goals:**
    * Replacing existing network-level firewalls (this is an application-level safety layer).
    * Inspecting encrypted payload content (focus is on destination validation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Agent Developer.
* **Primary Goal:** Prevent an AI agent from accessing the internal company wiki or cloud credentials even if the LLM is "tricked" into making the request.
* **The Happy Path (Tasks):**
    1. Developer configures MCP Any with an `egress_policy.yaml`.
    2. An LLM attempts to call a `fetch_url` tool with `url: "http://169.254.169.254/latest/meta-data/"`.
    3. MCP Any's Egress Shield intercepts the request before it reaches the upstream service.
    4. The Shield identifies the target as a restricted IP and blocks the request.
    5. MCP Any returns a "Security Policy Violation" error to the agent and logs the attempt in the Audit Log.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|Tool Call| Gateway[MCP Any Gateway]
        Gateway -->|Params| Middleware[Egress Shield Middleware]
        Middleware -->|DNS Lookup| DNS[DNS Resolver]
        DNS -->|Resolved IP| Middleware
        Middleware -->|Validate IP/Domain| Policy[Policy Engine]
        Policy -->|Allow/Deny| Middleware
        Middleware -->|Blocked| Gateway
        Middleware -->|Allowed| Upstream[Upstream API/Tool]
    ```
* **APIs / Interfaces:**
    * New Config Section: `security.egress_policy`.
    * Fields: `allowed_domains`, `denied_ips`, `allow_internal_networks` (default: false).
* **Data Storage/State:**
    * Policies loaded from YAML; violations logged to the central Audit Log.

## 5. Alternatives Considered
* **Network-Level Egress Control (Istio/Egress Gateway):** Rejected as the primary solution because it lacks tool-level context (e.g., knowing *which* tool triggered the request).
* **LLM-Based Validation:** Rejected because LLMs are susceptible to prompt injection and cannot reliably enforce security boundaries.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The shield follows the principle of "Default Deny" for all non-public or sensitive destinations.
* **Observability:** Every blocked request is tagged with the `Trace ID` and `Session ID` for security auditing.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation. Addressing OpenClaw-style SSRF vulnerabilities via application-level egress filtering.
