# Design Doc: Generic Filesystem Jail Middleware

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
Downstream MCP servers (e.g., `mcp-server-git`) frequently interact with the local filesystem. As evidenced by CVE-2026-27735, these servers often fail to properly validate user-supplied paths, leading to path traversal vulnerabilities. MCP Any must provide a "safety net" middleware that intercepts any tool arguments containing file paths and enforces strict jailing to authorized directories.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all tool calls and scan arguments for common path-like patterns (e.g., `path`, `filepath`, `root`, `target`).
    *   Resolve all paths against a whitelist of authorized "Repository Roots."
    *   Fail tool calls that attempt to access paths outside of authorized roots.
    *   Normalize paths to prevent `../` and symlink-based escapes.
*   **Non-Goals:**
    *   Reimplementing OS-level file permissions.
    *   Managing specific file-level access (it should be root-level).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious developer.
*   **Primary Goal:** Use a third-party Git MCP server without risk of the agent reading `/etc/passwd`.
*   **The Happy Path (Tasks):**
    1.  User configures MCP Any with a repository root: `/home/user/projects/my-app`.
    2.  Agent calls `git_add(path='../../etc/passwd')`.
    3.  Filesystem Jail Middleware intercepts the call, resolves the path, and detects it is outside the authorized root.
    4.  Middleware returns a security error to the agent, blocking the execution.

## 4. Design & Architecture
*   **System Flow:**
    - **Argument Interceptor**: A middleware that runs after policy checks but before tool execution.
    - **Path Resolver**: A utility that takes a raw string, resolves it to an absolute path, and checks it against the `authorized_roots` list.
    - **Recursive Scanning**: For complex tool arguments (JSON objects), the middleware recursively scans for keys that match "file path" patterns.
*   **APIs / Interfaces:**
    - Config entry: `filesystem_jail: { enabled: true, authorized_roots: ["/path/1", "/path/2"] }`
*   **Data Storage/State:** Authorized roots are stored in the main `config.yaml`.

## 5. Alternatives Considered
*   **Per-Tool Validation**: Hardcoding validation in every tool wrapper. *Rejected* as it is unscalable and prone to developer error.
*   **OS-Level Containers (Docker)**: Running every MCP server in a separate container. *Rejected* as it adds significant overhead and complexity.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This provides defense-in-depth against "Rug Pull" attacks and buggy third-party MCP servers.
*   **Observability:** All blocked path attempts should be logged as security events with the full attempted path.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
