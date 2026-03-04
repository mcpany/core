# Design Doc: Universal Tool Argument Sanitizer

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
The discovery of CVE-2026-0755 (RCE in `gemini-mcp-tool`) highlights a critical weakness: MCP adapters often trust their inputs. Many adapters pass arguments directly to shell commands or file APIs without adequate sanitization. MCP Any must provide a centralized defense-in-depth layer that sanitizes all tool arguments before they reach any adapter.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a middleware that recursively inspects all tool arguments.
    *   Detect and neutralize command injection patterns (e.g., `;`, `&&`, `|`, `` ` ``, `$()`).
    *   Prevent directory traversal in file-related tools (e.g., `../`).
    *   Support per-tool sanitization policies.
*   **Non-Goals:**
    *   Replacing tool-side validation (adapters should still validate their own inputs).
    *   Solving all possible application-level logic errors.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious AI Developer.
*   **Primary Goal:** Prevent an LLM from accidentally (or via prompt injection) executing malicious commands on the host machine through an MCP tool.
*   **The Happy Path (Tasks):**
    1.  The LLM generates a tool call for `gemini-mcp-tool` with a payload like `{"cmd": "ls ; rm -rf /"}`.
    2.  The Sanitizer middleware intercepts the call.
    3.  The Sanitizer detects the `;` character and flags the input.
    4.  The tool call is blocked, and an error is returned to the LLM: "Tool call blocked: Illegal characters in argument 'cmd'."

## 4. Design & Architecture
*   **System Flow:**
    - **Middleware Hook**: The Sanitizer sits early in the `mcpserver` request pipeline.
    - **Recursive Inspection**: It walks the JSON-RPC argument object recursively.
    - **Policy Engine**: Matches arguments against regex patterns or known "Safe" schemas.
*   **APIs / Interfaces:**
    - New middleware registration: `middleware.Register(Sanitizer)`
    - Configuration schema for `sanitization_rules`:
      ```yaml
      sanitization:
        enabled: true
        blocked_patterns: [";", "&&", "|", "`", "$(", "../"]
      ```
*   **Data Storage/State:** Stateless; policies are loaded from the configuration.

## 5. Alternatives Considered
*   **Adapter-Specific Fixes**: Wait for every adapter to fix their own 0-days. *Rejected* as it's too slow and prone to recurrence.
*   **WAF-style Filtering**: Use an external WAF. *Rejected* as it lacks the context of the MCP protocol and tool schemas.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust component, assuming that any input (even from an LLM) could be malicious.
*   **Observability:** All blocked calls must be logged with high severity to the Audit Log.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
