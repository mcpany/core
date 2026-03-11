# Design Doc: Tool Input Sanitization Middleware

**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
Recent vulnerabilities in MCP adapters (e.g., CVE-2026-0755 in `gemini-mcp-tool`) have shown that many MCP servers lack robust input validation. This allows attackers to perform command injection or other malicious acts by providing crafted tool arguments. As a "Universal Adapter," MCP Any must not just pass calls through, but act as a security shield that protects upstream servers and the host system.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a "Safe-by-Default" sanitization layer for all tool calls.
    *   Implement regex-based filtering of common command injection patterns (e.g., `;`, `&&`, `|`, `$(...)`).
    *   Support per-tool "Input Shape Attestation" where arguments are validated against strict schemas before being forwarded.
    *   Enable users to define custom "Blocklists" for specific tools or parameters.
*   **Non-Goals:**
    *   Perfectly solving all possible injection vectors (impossible without full semantic understanding).
    *   Replacing the security responsibility of the upstream server (it should still validate its inputs).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Ensure that no tool call, even from a compromised or poorly written subagent, can execute arbitrary system commands via shell injection.
*   **The Happy Path (Tasks):**
    1.  Architect enables the `SanitizationMiddleware` in the global MCP Any config.
    2.  An agent attempts to call a `run_script` tool with a malicious argument: `script_path="ls; rm -rf /"`.
    3.  `SanitizationMiddleware` intercepts the `tools/call` request.
    4.  The middleware detects the `;` injection character and the `rm -rf` pattern.
    5.  The call is blocked, and an error is returned to the agent (and logged in the security dashboard).
    6.  The upstream server never receives the malicious payload.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: The middleware sits at the beginning of the `Execution Pipeline`.
    - **Validation**:
        - **Schema Check**: First, validate arguments against the tool's JSON Schema.
        - **Regex Filtering**: Run a set of "Global Safety Regexes" against all string parameters.
        - **Contextual Analysis**: If the parameter name contains "path", "file", or "cmd", apply stricter "Path Traversal" and "Command Shell" filters.
*   **APIs / Interfaces:**
    - **Configuration**: A new `security.sanitization` block in `config.yaml` to define global and per-service rules.
*   **Data Storage/State:** None (Stateless middleware).

## 5. Alternatives Considered
*   **Upstream Patching**: Asking all MCP server authors to fix their code. *Rejected* as unscalable and outside our control.
*   **Full Virtualization**: Running every tool in a WASM or Docker sandbox. *Considered* (see `Detached Sandbox` feature) but may be too heavy for simple tools. Sanitization is a lighter, complementary layer.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust feature. It assumes the "Source" (Agent) and the "Target" (Upstream Server) are both potentially untrustworthy.
*   **Observability:** All blocked calls are logged with the "Blocked Reason" and the original payload for forensic analysis.

## 7. Evolutionary Changelog
*   **2026-03-11:** Initial Document Creation.
