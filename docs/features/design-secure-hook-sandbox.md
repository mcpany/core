# Design Doc: Secure Hook Sandbox

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of "auto-loading" agent configurations from repositories (pioneered by Claude Code and others), a new attack vector has emerged: malicious `CLAUDE.md` or `mcp.yaml` files that define automated hooks to execute arbitrary shell commands on the developer's machine. MCP Any must provide a secure execution environment for these hooks to prevent unauthorized system access while still allowing for useful automation.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Isolate hook execution from the host's primary shell environment.
    *   Implement a "Safety Prompt" or "Auto-Sanitization" for any hook involving sensitive commands (e.g., `rm -rf`, `curl | bash`).
    *   Provide a "Dry Run" mode for hooks to visualize side effects before execution.
    *   Standardize hook definitions across different agent frameworks (OpenClaw, Claude Code, Gemini CLI).
*   **Non-Goals:**
    *   Providing a full virtual machine for every hook (too heavy for local dev).
    *   Writing the hooks themselves (they are provided by the user or repository).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer cloning a new open-source project with an MCP Any configuration.
*   **Primary Goal:** Safely initialize the project's agentic tools without risking a shell-injection attack.
*   **The Happy Path (Tasks):**
    1.  User clones a repository containing a `.mcpany/config.yaml` with an `on_init` hook.
    2.  MCP Any detects the new configuration and identifies the hook: `npm install && ./setup_tools.sh`.
    3.  MCP Any parses the hook and identifies potentially dangerous patterns.
    4.  MCP Any prompts the user: "This repo wants to run a setup hook. [View Details] [Approve] [Decline]".
    5.  Upon approval, MCP Any executes the hook in a restricted sub-shell with a limited `PATH` and no access to sensitive environment variables (e.g., `STRIPE_API_KEY`).
    6.  The hook completes, and the tools are registered.

## 4. Design & Architecture
*   **System Flow:**
    - **Parsing**: A regex-based and AST-aware parser identifies commands and arguments in the hook string.
    - **Validation**: Hooks are checked against a "Denylist" of dangerous commands and a "Greylist" requiring explicit user approval.
    - **Execution**: Hooks are spawned using a dedicated `HookRunner` that utilizes OS-level primitives (like `unshare` on Linux or restricted `Job Objects` on Windows) to limit process capabilities.
*   **APIs / Interfaces:**
    - `mcpany_validate_hook(command: string) -> SafetyReport`
    - `mcpany_run_hook(command: string, sandbox_profile: string) -> HookResult`
*   **Data Storage/State:** Hook execution logs and user approval history are stored in the local SQLite audit log.

## 5. Alternatives Considered
*   **Docker-only Execution**: Running all hooks in Docker. *Rejected* as it requires Docker to be running and complicates access to the local source code being initialized.
*   **Strict Denylist**: Only allowing a small set of "safe" commands. *Rejected* because developers often need custom scripts that vary wildly.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Sandbox uses the "Principle of Least Privilege," only granting access to the specific directory and environment variables required for the hook.
*   **Observability:** All hook output (stdout/stderr) is captured and streamed to the MCP Any UI for real-time monitoring.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
