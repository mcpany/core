# Design Doc: Project-Scoped Security Guard
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent vulnerabilities (CVE-2026-0757, CVE-2025-59536) in agentic CLI tools like Claude Code have demonstrated that "opening a repository" is now a high-risk action. Malicious `mcp.config` files or repository hooks can lead to Remote Code Execution (RCE) and credential theft. MCP Any, acting as the universal gateway, must provide a "Security Sandbox" that isolates tool discovery and execution based on a verified project root, ensuring that untrusted repository-level configurations cannot compromise the host system.

## 2. Goals & Non-Goals
* **Goals:**
    * Restrict MCP server discovery to explicitly white-listed paths or the current verified project root.
    * Sanitize and validate all repository-local configuration files (`mcp.config`, `mcpany.yaml`) before ingestion.
    * Implement "Intent-Bound" execution where tools can only access files within the project boundary.
    * Provide a "Strict Mode" that disables all repository-local hooks by default.
* **Non-Goals:**
    * Replacing OS-level containerization (e.g., Docker). We are a middleware security layer.
    * Managing user authentication (handled by upstream providers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / Enterprise SRE
* **Primary Goal:** Safely explore a new open-source repository using an AI agent without risking machine takeover.
* **The Happy Path (Tasks):**
    1. User navigates to a new repository in their terminal.
    2. User starts `mcpany` with the `--project-root .` flag.
    3. MCP Any detects a local `mcp.config` but marks it as "Untrusted."
    4. MCP Any prompts the user (via HITL) to review the requested tools and permissions in the local config.
    5. User approves only the `linter` tool.
    6. The agent attempts to call a `shell` tool defined in the config; MCP Any blocks it because it wasn't approved and falls outside the "Safe Baseline."
    7. The agent reads a file; MCP Any verifies the path is within the project root.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [MCP Any Gateway] -> [Project Security Guard Middleware] -> [Policy Engine] -> [MCP Server]`
    The Middleware intercepts the `list_tools` and `call_tool` requests. It checks the `origin` of the tool definition.
* **APIs / Interfaces:**
    * New Config Section: `security_policies.project_isolation`
    * Internal Hook: `OnConfigLoad(path string)` - validates hashes and user approval state.
* **Data Storage/State:**
    * `.mcpany/trust.db`: A local SQLite database storing hashes of "Approved" project configurations.

## 5. Alternatives Considered
* **Full Dockerization**: Rejected due to high latency and complexity for local development workflows.
* **Global White-listing only**: Too restrictive for developers who need to use repo-specific tools.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses the "Principle of Least Privilege." Nothing is trusted by default.
* **Observability:** All blocked configuration attempts and tool calls are logged to the `audit_log` with a `SECURITY_VIOLATION` tag.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
