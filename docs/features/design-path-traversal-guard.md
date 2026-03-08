# Design Doc: Automated Path Traversal Guard

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
The disclosure of CVE-2026-28486 in the OpenClaw ecosystem demonstrated that AI agents are uniquely vulnerable to path traversal attacks when interacting with the web. Malicious sites can provide file paths (e.g., `../../etc/passwd`) that, if not properly sanitized by the tool or its host, can lead to host-level compromise. MCP Any, acting as the universal gateway, is perfectly positioned to provide a mandatory safety layer that sanitizes all path-based tool arguments before they reach the upstream MCP server.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all tool calls containing filesystem paths or archive operations.
    *   Enforce a "Rooted Sandbox" where all operations must stay within a designated directory.
    *   Provide automated sanitization of `..` and absolute paths.
    *   Integrate with the Policy Firewall to allow/deny specific path patterns.
*   **Non-Goals:**
    *   Modifying the upstream MCP server's code.
    *   Providing a full virtualized filesystem (this is a path sanitization layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Agent Developer.
*   **Primary Goal:** Prevent an agent from accidentally overwriting system files when using a "File Writer" or "Archive Extractor" tool.
*   **The Happy Path (Tasks):**
    1.  User enables the `PathTraversalGuard` in MCP Any config, specifying `sandbox_root: "/home/user/agent_files"`.
    2.  An agent receives a malicious instruction to "Save the report to `../../.ssh/authorized_keys`".
    3.  The agent calls the `write_file` tool with that path.
    4.  The `PathTraversalGuard` intercepts the call, detects the traversal attempt, and either:
        - Sanitizes it to `/home/user/agent_files/.ssh/authorized_keys` (if policy allows).
        - Rejects the call with a `SecurityViolation` error (default).
    5.  The upstream tool never sees the malicious path.

## 4. Design & Architecture
*   **System Flow:**
    - **Argument Inspection**: Middleware scans tool call JSON for common path-related keys (e.g., `path`, `filepath`, `dest`, `location`).
    - **Validation Engine**: Uses a "Strict Path Join" logic that resolves the final path and ensures it starts with the `sandbox_root`.
    - **Heuristic Detection**: Detects encoded traversal patterns (e.g., `%2e%2e%2f`).
*   **APIs / Interfaces:**
    - Configuration block in `mcp.yaml`:
      ```yaml
      middleware:
        path_guard:
          enabled: true
          sandbox_root: "./sandbox"
          allow_symlinks: false
          on_violation: "reject" # or "sanitize"
      ```
*   **Data Storage/State:** Stateless middleware, relies on per-call validation.

## 5. Alternatives Considered
*   **Tool-Specific Patching**: Fixing every MCP server. *Rejected* as it is unscalable and doesn't protect against new/unknown servers.
*   **OS-Level Sandboxing (Chroot/Docker)**: Running the whole MCP Any instance in a sandbox. *Considered* as a complementary measure, but path-level sanitization provides more granular control and better error reporting to the agent.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core "Safety" component. It must be "On by Default" in the Safe-by-Default strategy.
*   **Observability:** All path violations must be logged with high severity and visible in the Security Dashboard.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
