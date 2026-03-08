# Design Doc: Audit-First Argument Sanitization

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The "Prompt-to-RCE" vulnerability is a critical threat where LLMs generate tool arguments that include shell metasymbols or malicious payloads, which are then executed by vulnerable MCP servers. MCP Any, as the universal gateway, is uniquely positioned to intercept and sanitize these arguments before they reach the upstream tools. This feature implements a "Safety Schema" middleware that audits every tool call.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all `call_tool` requests.
    *   Validate arguments against a security-hardened schema.
    *   Strip or escape dangerous shell characters (e.g., `;`, `&`, `|`, `>`) if the tool is marked as "Shell-Sensitive."
    *   Provide a "Quarantine" mode where suspicious calls are sent for HITL approval.
*   **Non-Goals:**
    *   Rewriting the business logic of upstream MCP servers.
    *   Replacing the LLM's reasoning (we only sanitize the output).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer using Claude Code with a local Python tool.
*   **Primary Goal:** Prevent the agent from accidentally running `rm -rf /` via a malformed tool argument.
*   **The Happy Path (Tasks):**
    1.  Agent generates a tool call `run_python_script(code="import os; os.system('rm -rf /')")`.
    2.  MCP Any intercepts the call.
    3.  The Safety Schema identifies the `os.system` pattern as high-risk.
    4.  MCP Any blocks the execution and returns a "Security Violation" error to the agent, suggesting a safer alternative.

## 4. Design & Architecture
*   **System Flow:**
    - **Middleware Hook**: Injects into the `pkg/mcpserver` request pipeline.
    - **Pattern Matcher**: Uses a library of RegEx and AST-based filters to scan for injection patterns.
    - **Policy Engine**: References the `Policy Firewall` (Rego) to determine if a specific argument is allowed for the given agent context.
*   **APIs / Interfaces:**
    - `SecurityContext`: Added to the internal request object to track sanitization status.
    - `mcp_security_policy`: A new field in tool definitions allowing servers to opt-in to specific sanitization profiles (e.g., `profile: "shell-safe"`).
*   **Data Storage/State:**
    - Audit logs of blocked/sanitized calls stored in the local SQLite audit database.

## 5. Alternatives Considered
*   **LLM-Based Auditing**: Asking a second LLM to audit the call. *Rejected* due to latency and cost.
*   **Process Isolation**: Running every tool in a WASM sandbox. *Preferred long-term* but too complex for immediate "Audit-First" requirements.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core component of the Zero Trust execution boundary.
*   **Observability:** Every sanitization action must be logged with high fidelity for forensic analysis.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
