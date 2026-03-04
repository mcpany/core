# Design Doc: Global Path-Traversal Guard
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
Recent vulnerabilities in MCP servers (e.g., CVE-2026-27825 in mcp-atlassian) have shown that even "official" tools can have missing directory confinement, allowing remote attackers to write files to arbitrary paths. Instead of relying on every individual MCP server to be secure, MCP Any will implement a centralized "Guard" middleware that intercepts all tool calls containing path-like arguments and enforces strict sandbox boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect and sanitize file paths in tool arguments.
    * Enforce a per-service or per-session sandbox directory.
    * Block all `../` traversal attempts and symlinks pointing outside the sandbox.
* **Non-Goals:**
    * Implementing a full virtualized file system.
    * Changing the behavior of tools that do not take file paths as arguments.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Enterprise Architect.
* **Primary Goal:** Ensure that a compromised or buggy MCP server cannot write files to sensitive system directories (e.g., `/etc/` or `~/.ssh/`).
* **The Happy Path (Tasks):**
    1. Administrator configures a service in MCP Any with `sandbox_root: "/home/user/agent_workspace"`.
    2. An agent calls a tool `write_file(path="../../etc/passwd", content="...")`.
    3. The Global Path-Traversal Guard intercepts the call.
    4. The Guard detects the traversal attempt and the fact that the resolved path is outside the `sandbox_root`.
    5. The call is blocked, and an error is returned to the agent: `SecurityError: Path is outside of the authorized sandbox`.
    6. The event is logged in the Security Dashboard.

## 4. Design & Architecture
* **System Flow:**
    - **Argument Inspection**: Middleware parses the JSON-RPC request and identifies strings that look like filesystem paths (via regex or schema-based hints).
    - **Path Resolution**: Use `filepath.Abs` and `filepath.EvalSymlinks` to determine the true target.
    - **Boundary Check**: Verify that the resolved path starts with the configured `sandbox_root`.
* **APIs / Interfaces:**
    - Middleware hook in the tool execution pipeline.
    - Configuration schema extension: `sandbox_root` (string), `allow_symlinks` (boolean).
* **Data Storage/State:**
    - Configuration-driven; no runtime state required other than logging.

## 5. Alternatives Considered
* **OS-Level Sandboxing (Docker/jail)**: Excellent but adds significant overhead and complexity for local-first developer tools.
* **Manual Tool Review**: Too slow and error-prone given the rapid growth of the MCP marketplace.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements "Defense-in-Depth" by assuming the tool itself might be malicious or vulnerable.
* **Observability**: Integrates with the "Sandbox Perimeter Monitor" in the UI to show real-time blocks.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
