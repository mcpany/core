# Design Doc: Zero-Trust Egress Proxy for Tools

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
As agents become more autonomous, their ability to call tools that perform network requests increases. Today's market sync revealed multiple SSRF (Server-Side Request Forgery) vulnerabilities in the OpenClaw ecosystem, where malicious inputs could trick a tool into making requests to internal infrastructure or forbidden external domains. MCP Any must provide a secure, mediated egress layer for all tool-initiated network activity.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a mandatory Egress Proxy for all tools registered via MCP Any.
    *   Support per-tool network allow-lists (domain/IP/CIDR).
    *   Enable deep packet inspection (DPI) or header injection for tool requests.
    *   Provide audit logging for all tool egress traffic.
*   **Non-Goals:**
    *   Providing a general-purpose VPN for agents.
    *   Filtering traffic for the LLM itself (this is for tool-to-internet/internal traffic).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Ensure that a "Web Scraper" tool can only access public websites and never touch the company's internal Jenkins server at `10.0.0.50`.
*   **The Happy Path (Tasks):**
    1.  Architect configures the "Web Scraper" service in `mcp.yaml`.
    2.  Adds an `egress_policy` block: `allow: ["*.wikipedia.org", "*.github.com"]`.
    3.  When the agent triggers the scraper with a URL like `http://10.0.0.50/job/build`, the Egress Proxy blocks the request.
    4.  The agent receives a "Network Access Denied" error, and the attempt is logged in the Audit Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    - **Proxy Injection**: MCP Any injects `HTTP_PROXY` / `HTTPS_PROXY` environment variables into the tool's execution environment.
    - **Policy Enforcement**: The Egress Proxy intercepts all requests and matches them against the `egress_policy` defined for that specific tool/service.
    - **Transient Sandboxing**: For high-security tools, execution occurs in a network-isolated container where the *only* network interface is a bridge to the MCP Any Egress Proxy.
*   **APIs / Interfaces:**
    - `config.yaml` extension: `services.tool_name.egress_policy`
    - Middleware hook: `OnToolEgress(request)`
*   **Data Storage/State:** Policies are stored in-memory alongside the tool configuration; logs are persisted to the Audit database.

## 5. Alternatives Considered
*   **Host-Level Firewalls (iptables)**: Hard to manage across different OSs and dynamic tool sets.
*   **Application-Level Filtering**: Relying on the tool developer to implement security. *Rejected* as it violates the Zero Trust principle of the infrastructure layer.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Critical for preventing SSRF and data exfiltration from tools that handle user data.
*   **Observability:** Egress logs must be visible in the UI to allow developers to debug "blocked" legitimate requests.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
