# Design Doc: Isolated Workspace Middleware
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
As AI agents evolve from single-task assistants to multi-agent swarms, the need for parallel execution increases. Currently, running multiple agents in the same local directory leads to state collisions (e.g., git index locks, conflicting file writes). MCP Any needs a way to provide isolated execution environments for each agent session.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically provision isolated file system environments for agent sessions.
    * Support Git Worktrees as the primary isolation mechanism for git repositories.
    * Ensure each agent session has its own `CWD` (Current Working Directory).
* **Non-Goals:**
    * Full containerization (Docker/Podman) is out of scope for this initial middleware.
    * Managing network isolation between agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Run 5 parallel agents to refactor different modules of a large repository without merge conflicts or tool failures.
* **The Happy Path (Tasks):**
    1. Orchestrator sends a request to MCP Any to start a new "Isolated Session".
    2. MCP Any detects a git repository and creates a new git worktree in a temporary directory.
    3. MCP Any spawns the agent session with the temporary directory as the CWD.
    4. The agent performs its task and commits changes within its worktree.
    5. MCP Any provides a "Merge & Cleanup" tool to reconcile changes back to the main branch.

## 4. Design & Architecture
* **System Flow:**
    `Agent Session Request` -> `Workspace Middleware` -> `Git Worktree Manager` -> `Temporary Directory Allocation` -> `Tool Execution`.
* **APIs / Interfaces:**
    * `mcp_create_workspace(base_path string) -> workspace_id string, path string`
    * `mcp_delete_workspace(workspace_id string)`
* **Data Storage/State:**
    * A mapping of `session_id` to `worktree_path` is maintained in the MCP Any session store.

## 5. Alternatives Considered
* **Copy-on-Write Filesystems**: Too complex for local developer environments.
* **Simple Subdirectories**: Does not solve git index locking or branch management issues.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Workspaces should be restricted using existing MCP Any capability tokens to ensure agents cannot traverse outside their assigned temporary directory.
* **Observability**: Log workspace creation, path allocation, and cleanup events for debugging parallel runs.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation.
