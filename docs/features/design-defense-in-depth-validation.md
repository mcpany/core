# Design Doc: Defense-in-Depth Validation Middleware
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
Recent vulnerabilities in agent frameworks (e.g., OpenClaw) have shown that simple proxying of tool calls is insufficient. Attackers can leverage LLMs to perform SSRF, Path Traversal, and data exfiltration through tool parameters. MCP Any must provide a mandatory validation layer that inspects both tool inputs and outputs.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all tool call parameters against a strict schema.
    * Prevent SSRF by blocking tool calls that attempt to access restricted network ranges.
    * Prevent Path Traversal by sanitizing and validating file paths.
    * Inspect tool outputs for sensitive data patterns (e.g., API keys, PII) to prevent exfiltration.
* **Non-Goals:**
    * Implementing complex firewall rules (this is for the Policy Firewall).
    * Modifying tool behavior beyond validation/sanitization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer / Enterprise Admin
* **Primary Goal:** Ensure that tools accessed via MCP Any cannot be used as attack vectors for the local network or filesystem.
* **The Happy Path (Tasks):**
    1. Admin enables the `Defense-in-Depth` middleware in `config.yaml`.
    2. An LLM attempts to call a `fetch_url` tool with a local IP (e.g., `192.168.1.1`).
    3. The middleware detects the SSRF attempt based on the parameter's intent and blocks the call.
    4. The middleware logs the attempt and returns a security violation error to the LLM.

## 4. Design & Architecture
* **System Flow:**
    `LLM Tool Call` -> `MCP Any Gateway` -> **`Validation Middleware (Input)`** -> `Upstream MCP Server` -> `Tool Output` -> **`Validation Middleware (Output)`** -> `MCP Any Gateway` -> `LLM`
* **APIs / Interfaces:**
    * Middleware interface: `ValidateInput(params map[string]interface{}) error`
    * Middleware interface: `ValidateOutput(result map[string]interface{}) error`
* **Data Storage/State:**
    * Uses a local database (SQLite) for caching known malicious patterns and audit logs.

## 5. Alternatives Considered
* **Relying on Upstream Tools for Validation:** Rejected because many MCP servers are third-party and may not be well-secured.
* **Manual Tool Parameter Scoping:** Rejected as it's too cumbersome and error-prone for users.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Follows the "Verify Everything" principle.
* **Observability:** Detailed logging of all validation failures for security auditing.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
