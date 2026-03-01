# Design Doc: Secure Config Sandbox

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Recent vulnerabilities in Claude Code (CVE-2025-59536, CVE-2026-21852) have shown that local repository configuration files (e.g., `.mcpany/config.yaml`) can be a major attack vector for Remote Code Execution (RCE) and API token exfiltration. As agents increasingly work across diverse, untrusted repositories, MCP Any must provide a secure way to ingest local tools without exposing the host system.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Isolate repository-level configuration from the global system configuration.
    *   Require explicit user "Trust" for any local configuration file before execution.
    *   Prevent local configs from overriding sensitive global settings (e.g., Auth, Global Policy).
    *   Enable "Dry-Run" validation of local configs to check for malicious patterns (e.g., suspicious `ANTHROPIC_BASE_URL`).
*   **Non-Goals:**
    *   Replacing global configuration.
    *   Sandboxing the LLM itself (handled by model providers).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer.
*   **Primary Goal:** Safely use project-specific tools from a newly cloned GitHub repository.
*   **The Happy Path (Tasks):**
    1.  User clones a repository and runs `mcpany dev`.
    2.  MCP Any detects a local `.mcpany/config.yaml`.
    3.  MCP Any blocks the config and prompts the user: "Untrusted local config detected. Review and trust?".
    4.  User reviews the config (or runs `mcpany audit --local`).
    5.  User runs `mcpany trust .` to add the repo's hash to the local trust store.
    6.  MCP Any loads the local tools into a sandboxed session scope.

## 4. Design & Architecture
*   **System Flow:**
    - **Detection**: Filesystem watcher or path-based discovery for `.mcpany` directories.
    - **Trust Store**: A local SQLite/JSON store of `SHA256(config_path + content_hash)` that have been approved.
    - **Scoped Loading**: Local tools are loaded with a unique prefix (e.g., `local:repo_name:tool_name`) to prevent shadowing global tools.
*   **APIs / Interfaces:**
    - `mcpany trust [path]`: CLI command to approve a config.
    - `mcpany audit [path]`: CLI command to scan for known "malicious config" patterns.
*   **Data Storage/State:** Trust store managed in `~/.mcpany/trust.db`.

## 5. Alternatives Considered
*   **Automatic Sandboxing (Containers)**: Forcing all local tools into Docker. *Rejected* as default due to high overhead for simple scripts, but kept as an optional P1 feature.
*   **Environment Variable Blacklisting**: Only blocking specific keys. *Rejected* because attackers can use custom keys to achieve similar exfiltration (e.g., `PROXY_HOST`).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core purpose. Uses "No-Trust-by-Default" principle.
*   **Observability:** Log all attempts to load untrusted configs and the results of `audit` scans.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
