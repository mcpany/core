# Design Doc: Skill Workspace Sandbox Hardening

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
Local AI agents often operate in "Skill Workspaces"—directories where they can read/write files and execute code. Recent vulnerabilities (e.g., in OpenClaw) have shown that rogue or compromised subagents can escape these workspaces via symbolic link (symlink) attacks or path traversal (`../../`), gaining unauthorized access to the host filesystem. MCP Any, as the universal gateway, must provide a hardened, sandboxed environment for these local tool executions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Prevent subagents from accessing files outside their designated workspace.
    *   Detect and block symbolic link escapes.
    *   Provide "Workspace Chrooting" (filesystem-level isolation) for tool execution.
    *   Implement granular, path-based access control (Read/Write/Execute).
*   **Non-Goals:**
    *   Replacing full OS-level virtualization (like Docker), though we may leverage it.
    *   Managing host-level OS permissions (beyond the MCP Any process).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local Developer / Security-Conscious Agent User.
*   **Primary Goal:** Run a local "Code Refactoring" agent that is restricted to a single project directory without risk of it accessing `~/.ssh` or `/etc`.
*   **The Happy Path (Tasks):**
    1.  User configures a tool with a specific `workspace_root: "/path/to/project"`.
    2.  User sets `sandbox_mode: "strict"`.
    3.  The agent attempts to read `../../.env` via a tool call.
    4.  The `WorkspaceHardeningMiddleware` intercepts the call, detects the path traversal attempt, and blocks it.
    5.  An audit log is generated, and the agent receives a "Permission Denied" error.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: All filesystem-related tool calls are routed through the `Policy Firewall`.
    - **Validation**: The middleware resolves all paths to their absolute form and ensures they reside within the `workspace_root`.
    - **Symlink Check**: The system verifies that no symlink targets point outside the root.
*   **APIs / Interfaces:**
    - **Config**: New `sandbox` block in service/tool definitions.
    - **Hooks**: Filesystem hooks (`fs_read`, `fs_write`, `fs_exec`) for the Policy Engine.
*   **Data Storage/State:** Workspace definitions and active sandbox states are stored in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Full Dockerization**: Running every tool in a separate container. *Rejected* for local use cases due to high latency and resource overhead.
*   **OS User Isolation**: Running MCP Any as a low-privileged user. *Rejected* because it doesn't solve the problem of one tool compromising another tool's workspace within the same process.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Follows the principle of least privilege. Workspaces are "Closed by Default."
*   **Observability:** All sandbox violations are logged to the `Security Dashboard` with high severity.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
