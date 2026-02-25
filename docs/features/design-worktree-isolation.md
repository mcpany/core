# Design Doc: Worktree Isolation Provider
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents (like Claude Code and OpenClaw) perform increasingly complex tasks involving multi-file edits and git operations, they often pollute the user's primary working directory. This leads to accidental commits of "work-in-progress" junk, merge conflicts between parallel agent tasks, and potential data loss.

MCP Any needs to provide a standardized way to isolate these agent "thoughts" and "actions" into dedicated Git worktrees. This ensures that agents can operate in a clean environment that mirrors the main repo without affecting the main branch until the task is explicitly "merged" or "approved".

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically create a lightweight Git worktree for every new agent session/task.
    *   Synchronize the worktree state back to the main branch upon task completion (with user approval).
    *   Automatically clean up (delete) worktrees after session expiry or task finalization.
    *   Provide a "Worktree Tool" that agents use to manage their isolated environment.
*   **Non-Goals:**
    *   Replacing full containerization (Docker). This is specifically for filesystem/git-heavy tasks.
    *   Managing non-git projects (initially).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Senior Software Engineer using a swarm of agents for a large-scale refactor.
*   **Primary Goal:** Run 3 subagents in parallel to refactor 3 different modules without them interfering with each other's git state or the main `HEAD`.
*   **The Happy Path (Tasks):**
    1.  User starts a "Refactor Session" via MCP Any.
    2.  MCP Any identifies the project is a Git repo and triggers the `Worktree Isolation Provider`.
    3.  A new worktree is created in a hidden `.mcpany/worktrees/[session-id]` directory.
    4.  Subagents are launched, each receiving a `base_path` pointing to their specific worktree.
    5.  Agents perform edits, run tests, and commit locally within the worktree.
    6.  Once the task is done, the user reviews a unified diff of all worktrees.
    7.  User clicks "Approve", and MCP Any merges the worktree changes into the main branch and deletes the worktree.

## 4. Design & Architecture
*   **System Flow:**
    `Agent Request` -> `Session Middleware` -> `Worktree Provider` -> `git worktree add` -> `Tool Execution`
*   **APIs / Interfaces:**
    *   `session.isolation = "worktree"` (Configuration flag)
    *   `mcp_worktree_status`: Tool to check current isolation state.
    *   `mcp_worktree_commit`: Tool to commit changes within the isolated environment.
*   **Data Storage/State:**
    *   Worktree metadata stored in the Shared KV Store (Blackboard).
    *   Physical worktrees stored in `.mcpany/worktrees/`.

## 5. Alternatives Considered
*   **Docker Isolation**: Too heavy for simple file edits; requires Docker to be installed and running.
*   **Branch-based Isolation**: Pollutes the remote/local branch list; harder to manage "orphan" branches.
*   **Copy-on-Write Filesystems**: Not portable across Windows/macOS/Linux.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):**
    *   The `Policy Firewall` must restrict the agent to only its assigned worktree path.
    *   Prevent `../` traversal out of the worktree.
*   **Observability:**
    *   UI Timeline shows worktree creation, commits, and deletion events.
    *   Disk usage monitoring for the `.mcpany/worktrees` directory.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
