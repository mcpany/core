# Design Doc: Hook Approval Protocol (Safe-Hooks)

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
Recent vulnerabilities in Claude Code (February 2026) demonstrated that automated hooks defined in project configuration files can be used as a vector for Remote Code Execution (RCE). When a developer clones a repository and initializes an agentic tool, malicious hooks can execute arbitrary commands without explicit consent. MCP Any must provide a robust defense-in-depth mechanism to ensure that no automated command is executed without verifiable user intent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Requirement for explicit user approval before any newly discovered hook is executed.
    *   Support for cryptographic signing of "trusted hooks" by known identities.
    *   Maintain an audit log of all hook executions and their approval status.
    *   Provide a "Safe-Hooks" mode where unauthenticated hooks are blocked by default.
*   **Non-Goals:**
    *   Scanning the hook commands for malicious intent (this is a sandbox/policy concern).
    *   Replacing traditional CI/CD pipelines.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer cloning a new open-source project.
*   **Primary Goal:** Safely initialize the project's MCP environment without running rogue scripts.
*   **The Happy Path (Tasks):**
    1.  User clones a repository containing an `mcpany.yaml` with a `post_init` hook.
    2.  User runs `mcpany start`.
    3.  MCP Any detects the new hook and pauses initialization.
    4.  The CLI/UI displays: "New Hook Detected: `rm -rf /` (Source: mcpany.yaml). Do you want to allow this? [y/N/Always for this Repo]".
    5.  User denies the hook. MCP Any continues initialization without executing the dangerous command.

## 4. Design & Architecture
*   **System Flow:**
    - **Hook Discovery**: During configuration loading, the `HookRegistry` identifies all command-based hooks.
    - **Integrity Check**: The system computes a SHA256 hash of the command and its execution context.
    - **Approval Verification**: The `ApprovalStore` is queried for the hash. If missing or "pending," the execution is suspended.
    - **Out-of-Band (OOB) Approval**: A notification is sent to the UI/CLI.
*   **APIs / Interfaces:**
    - `POST /v1/hooks/approve`: Endpoint to manually approve a hook hash.
    - `GET /v1/hooks/pending`: List hooks awaiting approval.
*   **Data Storage/State:**
    - `hooks.db`: SQLite table storing `hook_hash`, `command_string`, `source_path`, `status` (Approved/Denied/Pending), and `timestamp`.

## 5. Alternatives Considered
*   **Sandboxing only**: Running hooks in a restricted container. *Rejected* as even sandboxed hooks can leak secrets via environment variables.
*   **Static Analysis**: Attempting to "lint" hooks for danger. *Rejected* as it is trivial to bypass via obfuscation.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature directly addresses the "Configuration Poisoning" threat model.
*   **Observability:** All hook execution attempts (including blocked ones) must be recorded in the main Audit Log.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
