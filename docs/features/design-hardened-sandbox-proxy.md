# Design Doc: Hardened Sandbox Proxy Middleware
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
With the discovery of multi-layer SSRF in OpenClaw and RCE vulnerabilities in Claude Code, it is clear that AI agents often interact with tools that can be manipulated to perform unauthorized network requests or filesystem operations. MCP Any, as the universal adapter, must ensure that any tool call passing through it is safe, regardless of whether the underlying MCP server is trusted.

This middleware acts as a security interceptor for all tool requests and responses, providing a hardened perimeter around tool execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all tool input arguments for SSRF patterns (internal IPs, localhost, protocol smuggling).
    * Prevent path traversal by sanitizing file paths in tool arguments.
    * Scrub tool outputs to prevent sensitive data leakage or secondary injection.
    * Provide a "Local-Only" enforcement mode for tools that shouldn't access the network.
* **Non-Goals:**
    * Implementing a full network firewall (relies on host-level security).
    * Modifying the business logic of the underlying tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Run a community-contributed MCP server (e.g., a web scraper) without risking an SSRF attack on the internal network.
* **The Happy Path (Tasks):**
    1. User adds a third-party MCP server to `mcpany.yaml`.
    2. User enables the `hardened_proxy` middleware for that service.
    3. An agent attempts to call a `scrape_url` tool with an internal IP (e.g., `http://192.168.1.1/admin`).
    4. The Hardened Sandbox Proxy detects the internal IP range.
    5. The request is blocked, and an error is returned to the agent without calling the upstream tool.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any Gateway -> [Policy Firewall] -> [Hardened Sandbox Proxy] -> Upstream Tool`
* **APIs / Interfaces:**
    * Middleware configuration in `mcpany.yaml`:
      ```yaml
      services:
        risky-tool:
          middleware:
            - type: hardened_proxy
              allow_internal: false
              allowed_schemes: [http, https]
              path_restrictions: [/tmp/sandbox]
      ```
* **Data Storage/State:** Stateless validation based on request/response payloads.

## 5. Alternatives Considered
* **Docker Isolation:** Highly secure but introduces significant overhead and complexity for local-first tools.
* **OS-Level Sandboxing (gVisor/Firecracker):** Best-in-class security but difficult to distribute as a lightweight CLI/Gateway.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The proxy implements a deny-by-default strategy for internal network ranges.
* **Observability:** All blocked requests are logged with high-fidelity "Actionable Errors" explaining why the request was deemed unsafe.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
