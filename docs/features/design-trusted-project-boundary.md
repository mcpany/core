# Design Doc: Trusted Project Boundary Isolation

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
Recent vulnerabilities in Claude Code (CVE-2026-21852) demonstrated that agents are susceptible to "Configuration-as-Attack-Vector." By simply opening a malicious repository, an agent might ingest `.env` files or project settings that redirect its API traffic to an attacker-controlled proxy or execute malicious hooks. MCP Any needs a mechanism to isolate and "Trust" project boundaries before honoring local configurations.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a "Trust Prompt" or "Trust Registry" for project directories.
    *   Ignore all project-local `.mcpany/config` and `.env` files by default for untrusted directories.
    *   Provide a "Safe Mode" for tool execution within untrusted boundaries (e.g., restricted filesystem access).
    *   Persist "Trust" state across sessions securely.
*   **Non-Goals:**
    *   Full containerization of the agent (handled by external runtimes like Docker).
    *   Scanning repositories for malware (this is a configuration security layer, not an antivirus).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious developer cloning a new open-source project.
*   **Primary Goal:** Use MCP Any tools on the project without risking credential theft or RCE from malicious project settings.
*   **The Happy Path (Tasks):**
    1.  User `cd`s into a newly cloned repository and runs an MCP Any-enabled tool.
    2.  MCP Any detects a local `.mcpany/config` but notices the directory is not in the `trusted_workspaces.json`.
    3.  MCP Any starts with global defaults only and displays a warning: "Untrusted Project: Local configurations ignored. Run `mcpany trust` to enable."
    4.  User reviews the local config and runs `mcpany trust`.
    5.  Subsequent executions use the project-local configurations.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Phase**: On startup or directory change, the `BoundaryManager` identifies the nearest `.mcpany/` or `.git/` root.
    - **Verification Phase**: The manager checks the `~/.mcpany/trusted_workspaces.json` (keyed by absolute path and directory hash).
    - **Policy Enforcement**: If untrusted, the `ConfigLoader` skips project-local files. If trusted, it merges them with global settings.
*   **APIs / Interfaces:**
    - `mcpany trust [path]`: Adds a path to the trust registry.
    - `mcpany untrust [path]`: Removes a path.
    - Middleware Hook: `OnProjectLoad(context)`
*   **Data Storage/State:**
    - `~/.mcpany/trusted_workspaces.json`: Stores `{"path": "/abs/path", "hash": "sha256_of_root_metadata"}`.

## 5. Alternatives Considered
*   **Prompting on every file read**: Too much friction for the user.
*   **Always ignoring project configs**: Breaks the "useful by default" goal for complex projects that require specific tool setups.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This directly mitigates the "Malicious Repository" attack vector.
*   **Observability:** The `mcpany status` command should indicate if the current workspace is trusted.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
