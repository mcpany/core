# Design Doc: Quarantine-First Configuration Controller
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
Recent critical vulnerabilities in Claude Code (CVE-2025-59536) have demonstrated the risk of "Auto-loading" configuration files from untrusted sources. When a repository is cloned, malicious actors can inject automated hooks or override MCP server configurations to execute arbitrary code or exfiltrate sensitive data.

MCP Any needs a "Quarantine-First" mechanism that prevents any imported or auto-discovered configuration (e.g., `.mcp/config.yaml`, `.claude/config.json`) from being active until a verified user has manually reviewed and "Released" it.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all newly discovered or imported configurations.
    * Maintain a "Quarantined" state for configurations that have not been user-verified.
    * Prevent any tools, hooks, or transport listeners from these configurations from starting.
    * Provide a secure workflow for a human user to review and release configurations.
* **Non-Goals:**
    * Automating the "Release" of configurations (defeats the security purpose).
    * Fixing or sanitizing malicious configurations (responsibility lies with the user/author).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer working with multiple agent-orchestrated repositories.
* **Primary Goal:** Safely clone and work with a new repository without risking silent hook execution.
* **The Happy Path (Tasks):**
    1. User clones a repository containing a `.mcp/config.yaml`.
    2. MCP Any detects the new configuration file.
    3. MCP Any places the config in a "Quarantined" state and notifies the user (via UI/CLI).
    4. User opens the MCP Any Dashboard to the "Security Quarantine" section.
    5. User reviews the proposed changes (e.g., "This repo wants to add a pre-hook that runs `rm -rf /`").
    6. User selectively "Releases" or "Rejects" the configuration components.
    7. Only released components are activated in the runtime.

## 4. Design & Architecture
* **System Flow:**
    - **Config Watcher**: Monitors the filesystem for new or modified config files.
    - **Quarantine Manager**: Stores metadata about quarantined configs in the internal KV Store.
    - **Admission Controller**: A gatekeeper in the service loading pipeline that checks the quarantine status before initializing any service or hook.
* **APIs / Interfaces:**
    - `GET /api/v1/quarantine`: List all quarantined configurations and their components.
    - `POST /api/v1/quarantine/release`: Mark a specific config or component as released.
    - `POST /api/v1/quarantine/reject`: Purge a quarantined configuration.
* **Data Storage/State:** Quarantined configurations are stored as "Inactive" records in the internal SQLite DB with a `status` field (`quarantined`, `released`, `rejected`).

## 5. Alternatives Considered
* **Sandboxing Hooks**: Running hooks in an isolated Docker container. Rejected because it's complex and doesn't solve the "Secret Exfiltration" problem if the container has network access.
* **Signature-Only**: Only allowing configs signed by a trusted authority. Rejected because it's too restrictive for local development and shared community repos.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Release" action itself must be protected by a secure session (e.g., local MFA or session token) to prevent a malicious agent from releasing its own malicious config.
* **Observability:** All quarantine/release actions are logged to the Audit Log with user identity and timestamp.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
