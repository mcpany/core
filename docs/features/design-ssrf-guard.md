# Design Doc: SSRF Guard Middleware
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
The March 2026 OpenClaw SSRF (GHSA-56f2-hvwg-5743) incident revealed a systemic vulnerability in AI agent tool calls: an agent can be coerced into using a "benign" tool (like an image fetcher) to probe or attack internal infrastructure. Since many MCP servers are thin wrappers around existing libraries, they often lack robust network isolation. MCP Any, as the universal gateway, must provide a centralized, protocol-level protection layer.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all tool arguments that appear to be URLs.
    *   Enforce a strict blocklist of private (RFC1918), link-local, and cloud metadata IPs (e.g., `169.254.169.254`).
    *   Support hostname allow-listing for sensitive environments.
    *   Provide cryptographic "Fetch Attestation" for allowed requests.
*   **Non-Goals:**
    *   Rewriting the underlying tool's fetch logic (we act as a gatekeeper, not the implementer).
    *   Monitoring network traffic at the OS level (this is middleware, not a kernel firewall).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Engineer at a FinTech firm.
*   **Primary Goal:** Ensure that no agent can access internal microservices via the `fetch_web_content` or `image_to_text` tools.
*   **The Happy Path (Tasks):**
    1.  Engineer enables `ssrf_guard` in the `mcpany` config.
    2.  An agent tries to call `image_to_text(url="http://169.254.169.254/latest/meta-data/")`.
    3.  SSRF Guard intercepts the call, identifies the target as a restricted cloud metadata IP, and blocks the execution.
    4.  A security alert is logged in the `mcpany` Audit Log with the blocked target and the agent's current intent context.

## 4. Design & Architecture
*   **System Flow:**
    - **Argument Interception**: The middleware scans JSON-RPC tool call arguments for strings matching common URL patterns.
    - **DNS Resolution Check**: The middleware performs a "Safe DNS Lookup" (pinning) to ensure a hostname doesn't resolve to a private IP after validation (mitigating TOCTOU).
    - **Policy Evaluation**: The target IP/Host is checked against the `PolicyEngine`.
    - **Gatekeeping**: If the target is blocked, the tool call is rejected with a `403 Forbidden` error before it reaches the MCP server.
*   **APIs / Interfaces:**
    - **Config Schema**: `ssrf_guard: { allowed_hosts: ["*.example.com"], block_private_ips: true }`
*   **Data Storage/State:** Cached DNS resolutions for performance.

## 5. Alternatives Considered
*   **Tool-Level Hardening**: Fixing every individual MCP server. *Rejected* as it's impossible to maintain across the entire ecosystem.
*   **Egress Firewall**: Using `iptables` or `eBPF`. *Rejected* as it requires root privileges and doesn't have the "Agent Intent" context.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Critical. This is a "Defense in Depth" layer.
*   **Observability:** All blocked fetch attempts are logged as high-priority security events.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
