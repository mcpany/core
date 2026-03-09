# Design Doc: Agent Worktree Isolation Provider

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
Modern AI agents (e.g., Claude Code, OpenClaw) increasingly perform autonomous file operations and command execution. Running these directly on a user's primary filesystem or in a shared environment is risky. Following the "isolation-first" pattern seen in Claude Code 2.1.50, MCP Any needs a standardized way to provide agents with ephemeral, isolated workspaces (e.g., git worktrees, Docker volumes, or named pipes) that are managed by the infrastructure layer.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide standard MCP tools for requesting and destroying isolated workspaces.
    *   Support multiple isolation providers (Git Worktree, Docker, TempFS).
    *   Enforce "Workspace Trust" policies via the Policy Firewall.
    *   Standardize lifecycle hooks (`OnCreate`, `OnDestroy`) for workspace initialization.
*   **Non-Goals:**
    *   Building a new container orchestration system.
    *   Providing long-term persistent storage (workspaces are ephemeral).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Autonomous DevOps Agent.
*   **Primary Goal:** Safely refactor a library without risking corruption of the main repository branch.
*   **The Happy Path (Tasks):**
    1.  Agent calls `request_isolated_workspace(type="git-worktree", base_ref="main")`.
    2.  MCP Any validates the request against policy and calls the Git provider.
    3.  A new worktree is created; MCP Any returns the absolute path to the agent.
    4.  Agent performs refactoring within the worktree.
    5.  Agent calls `destroy_isolated_workspace(id="worktree_123", sync_changes=true)`.
    6.  MCP Any merges changes if requested and prunes the worktree.

## 4. Design & Architecture
*   **System Flow:**
    - **Request**: Agent tool call -> `IsolationMiddleware` -> `ProviderRegistry` -> `Git/Docker Provider`.
    - **Policy Check**: Every request is intercepted by the `Policy Firewall` to ensure the agent has `fs:isolate` capabilities.
    - **Mapping**: MCP Any maintains a mapping of `WorkspaceID -> HostPath` to prevent directory traversal attacks.
*   **APIs / Interfaces:**
    - `mcp_any_isolate_create(provider_type, config)`
    - `mcp_any_isolate_destroy(id, options)`
    - `mcp_any_isolate_list()`
*   **Data Storage/State:** Workspace metadata is stored in the `Shared KV Store` to ensure cleanup even if the server restarts.

## 5. Alternatives Considered
*   **User-Managed Isolation**: Forcing users to manually set up environments. *Rejected* because it creates too much friction for "autonomous" workflows.
*   **Native Agent Isolation**: Letting every agent framework implement its own isolation. *Rejected* because it leads to fragmented security policies and inconsistent cleanup.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Workspaces are isolated from the host `ENV`. MCP Any redacts host-level secrets unless explicitly passed during creation.
*   **Observability:** Monitor workspace disk usage and lifecycle events in the UI.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
