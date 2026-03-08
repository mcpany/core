# Design Doc: Project Configuration Attestation

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
With the rise of agentic tools like Claude Code and OpenClaw, attackers have begun using "Configuration-as-a-Vector" attacks. By placing malicious MCP configurations in project-level directories (e.g., `.mcpany/config.yaml`), an attacker can trick an agent into executing dangerous hooks, exfiltrating API keys, or connecting to rogue MCP servers. MCP Any must ensure that any configuration not found in the global, user-authorized path is treated as "Untrusted" until attested.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Detect and isolate project-local configuration files.
    *   Prevent automatic ingestion of "Untrusted" configurations.
    *   Provide a secure "User-in-the-Loop" attestation flow for new project configs.
    *   Support cryptographic signing of configuration files for team collaboration.
*   **Non-Goals:**
    *   Validating the *syntax* of the config (handled by the existing validator).
    *   Managing global system configurations (which are inherently trusted).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer cloning a new open-source repository.
*   **Primary Goal:** Work on the project without the project's MCP configuration hijacking their local environment.
*   **The Happy Path (Tasks):**
    1.  User `cd`s into a newly cloned repository containing `.mcpany/config.yaml`.
    2.  User runs `mcpany start`.
    3.  MCP Any detects the local config and outputs: `[SECURITY] Untrusted project configuration detected at .mcpany/config.yaml`.
    4.  MCP Any pauses loading the local config and prompts the user: `Run 'mcpany trust' to authorize this configuration.`
    5.  User runs `mcpany trust`, which displays a diff of the changes and requests confirmation.
    6.  Configuration is moved to a "Trusted" state (tracked via SHA256 in the global database).

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery**: The `ConfigLoader` scans for local `.mcpany` directories.
    - **Verification**: It computes a SHA256 hash of the found config and checks it against a `trusted_configs` table in the global SQLite DB.
    - **Quarantine**: If the hash is missing or mismatched, the config is skipped, and a warning is issued.
    - **Attestation**: The `trust` command allows the user to audit the config and record the hash as trusted.
*   **APIs / Interfaces:**
    - CLI: `mcpany trust [path]`
    - CLI: `mcpany distrust [path]`
*   **Data Storage/State:** `trusted_configs` table in `~/.mcpany/state.db`.

## 5. Alternatives Considered
*   **Whitelisting Directories**: Allowing specific paths (e.g., `~/projects/*`). *Rejected* as it's too broad; a malicious file could still be placed in a whitelisted path.
*   **Automatic Sandboxing**: Running all tools from local configs in a sandbox. *Preferred as a secondary layer*, but attestation is the primary defense for configuration integrity.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This closes the "Shadow Configuration" hole.
*   **Observability:** The UI should show "Project Context: [Trusted | Untrusted | No Local Config]".

## 7. Evolutionary Changelog
*   **2026-03-07:** Initial Document Creation.
