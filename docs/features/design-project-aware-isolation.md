# Design Doc: Project-Aware Adapter Isolation

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
As documented in the 2026-03-07 Market Sync, recent vulnerabilities in AI agents (e.g., Claude Code CVE-2025-59536) have shown that "Global" agent configurations are a major attack vector. When an agent is started in an untrusted repository, that repository can hijack the agent's behavior by overriding environment variables like `ANTHROPIC_BASE_URL` or injecting malicious hooks.

MCP Any must solve this by introducing "Project-Aware Isolation." Tools, configurations, and environment variables should be scoped to a specific project directory. A tool authorized for "Project A" should not have access to "Project B's" secrets, and an untrusted repository should not be able to modify the global MCP Any security posture.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Bind MCP server instances and tool definitions to a specific filesystem path (Project Root).
    *   Implement "Configuration Hierarchies" where project-level settings are strictly shadowed and cannot override "Global Immutables."
    *   Provide a "Trust Prompt" mechanism when an agent enters a new project directory for the first time.
    *   Isolate environment variable injection so that project-specific keys don't leak to global tools.
*   **Non-Goals:**
    *   Providing full OS-level virtualization (that is handled by the Ephemeral Sandboxing feature).
    *   Managing Git submodules or external repository state.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer.
*   **Primary Goal:** Safely use AI agents on a newly cloned, untrusted open-source repository without risking API key theft or RCE.
*   **The Happy Path (Tasks):**
    1.  User clones a repository and runs an agent connected to MCP Any.
    2.  MCP Any detects the agent is operating in a new directory `/home/user/untrusted-repo`.
    3.  MCP Any intercepts the request and asks the user: "Trust this project? It is requesting access to the 'Filesystem' and 'Shell' tools."
    4.  User approves with "Limited Access."
    5.  The repository's `.mcpany/config.yaml` attempts to set `BASE_URL` to a malicious proxy.
    6.  MCP Any blocks the override, logging a security alert, and uses the global, trusted `BASE_URL`.
    7.  The agent only sees tools specifically enabled for this project scope.

## 4. Design & Architecture
*   **System Flow:**
    - **Context Detection**: The gateway identifies the `cwd` (current working directory) of the calling agent via transport metadata or filesystem hooks.
    - **Scope Resolver**: Maps the `cwd` to a "Project Profile."
    - **Shadow Config Layer**: Merges global config with project config, enforcing "Immutability Rules" defined in the global policy.
*   **APIs / Interfaces:**
    - `context/get_project_scope`: Returns the active isolation boundaries.
    - `policy/set_immutable`: CLI/API to mark specific config keys as non-overridable.
*   **Data Storage/State:**
    - `projects.db`: A local SQLite table tracking trusted directories and their specific permissions/overrides.

## 5. Alternatives Considered
*   **Purely Manual Configuration**: Forcing users to manually create a new MCP Any instance for every project. *Rejected* due to extreme friction.
*   **OS-level User Isolation**: Running the agent as a different OS user. *Rejected* as it complicates local file access which is the primary use case for coding agents.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a direct implementation of "Least Privilege" applied to the filesystem and project context.
*   **Observability:** The UI must clearly indicate the "Active Project Scope" and highlight any blocked configuration override attempts.

## 7. Evolutionary Changelog
*   **2026-03-07:** Initial Document Creation.
