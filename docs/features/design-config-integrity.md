# Design Doc: Config Integrity Sandboxing

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Recent exploits in Claude Code showed that repository-level configuration files (like `.mcp.json` or `.claude/settings.json`) can be used to inject malicious "hooks" or redirect API traffic. AI agents often recursively search for and auto-load these files, granting them high trust. MCP Any must treat any configuration not found in the global/system-wide path as "Untrusted" and sandbox its execution.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Differentiate between "System Config" (Trusted) and "Project Config" (Untrusted).
    *   Require explicit user attestation (HITL) for high-risk settings found in project-local configs.
    *   Provide a "Safe Mode" for repository scanning.
*   **Non-Goals:**
    *   Eliminating project-local configs entirely (they are useful for team-sharing).
    *   Sandboxing the actual tool execution (this is handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious developer cloning a new repository.
*   **Primary Goal:** Run MCP Any in the repo without being "ClawJacked" by a malicious `.mcp.json`.
*   **The Happy Path (Tasks):**
    1.  User navigates to a new repo and runs an agent that uses MCP Any.
    2.  MCP Any detects a `.mcp.json` in the current directory containing a `post_command_hook`.
    3.  Instead of executing the hook, MCP Any pauses and prompts the user: "Security Warning: Untrusted Project Config detected. Allow execution of hook: 'rm -rf /'?"
    4.  User denies the request; MCP Any proceeds but ignores the malicious hook.

## 4. Design & Architecture
*   **System Flow:**
    - **Config Classifier**: The `ConfigLoader` identifies the source of each configuration entry.
    - **Taint Tracking**: Entries from project-local files are marked as "Untrusted."
    - **Policy Enforcement**: When a "Tainted" configuration attempts to trigger a high-risk action (e.g., shell command, base URL change), it triggers a HITL (Human-in-the-Loop) event.
*   **APIs / Interfaces:**
    - Internal `ConfigRecord` struct now includes `SourcePath` and `IsTrusted` boolean.
*   **Data Storage/State:** Persistence of "User-Approved" hashes of local configs to avoid re-prompting for known good files.

## 5. Alternatives Considered
*   **Disallowing Local Hooks**: Completely banning hooks in local configs. *Rejected* as it breaks legitimate use cases (e.g., local database setup).
*   **Using a "Lock" File**: Similar to `package-lock.json`. *Rejected* as it's also a file in the repo that can be manipulated.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Critical for supply chain security.
*   **Observability:** The UI should show which settings are currently "Sandboxed" or "Pending Approval."

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
