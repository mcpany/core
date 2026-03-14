# Design Doc: Settings Injection Guard
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
Recent security research (CVE-2026-25725) has highlighted a critical vulnerability where AI agents automatically ingest configurations from project-local files (e.g., `.claude/settings.json`). Malicious actors can weaponize these files to inject "auto-execute" hooks, redirect API base URLs for exfiltration (Rug Pull), or lower security sandbox levels. MCP Any must provide an active interception and validation layer to ensure these configuration files match an attested baseline before they reach the agent runtime.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept read/write access to project-local agent configuration files.
    * Validate configuration keys against a "Safe Baseline" (e.g., no base URL redirection, no unauthorized hooks).
    * Require explicit user attestation (MFA) for any changes to security-critical settings.
    * Provide a real-time monitor for configuration drift.
* **Non-Goals:**
    * Managing the agent's internal state (only the external configuration files).
    * Providing a general-purpose filesystem firewall (scope is limited to agent-adjacent configs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a cloned repository from exfiltrating API keys via a malicious `.claude/settings.json` file.
* **The Happy Path (Tasks):**
    1. The developer clones a repository and runs an AI agent (e.g., Claude Code).
    2. The agent attempts to read `.claude/settings.json`.
    3. MCP Any's **Settings Injection Guard** intercepts the read request.
    4. The Guard compares the file's content against the user's "Global Safety Policy."
    5. The Guard detects an unauthorized `baseUrl` override and a malicious `post-execution-hook`.
    6. MCP Any blocks the ingestion and alerts the user via the UI/CLI.
    7. The user reviews the diff and "Quarantines" the malicious settings.
    8. The agent is provided with a "Sanitized" version of the config.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Runtime] <-> [Settings Injection Guard (Proxy/FUSE)] <-> [Project Filesystem]`
* **APIs / Interfaces:**
    * `ConfigInterceptor`: Internal service that hooks into filesystem events for `.json`, `.yaml`, and `.toml` files in `.agent/` or `.claude/` directories.
    * `/v1/guard/attest`: Endpoint for users to approve or reject configuration changes.
* **Data Storage/State:**
    * Stores "Known Good" hashes of project configurations in the SQLite Blackboard.
    * Uses `Content-Addressable Config (CAC)` for integrity verification.

## 5. Alternatives Considered
* **Static Scanning (Pre-commit):** Rejected because it doesn't protect against runtime changes or "TOCTOU" (Time-of-Check to Time-of-Use) races.
* **Read-Only Enforced Directories:** Rejected as it breaks legitimate agent functionality (e.g., agents needing to update their own session state).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Guard itself must be hardware-attested to prevent tampering.
* **Observability:** All interception events and attestation results are logged to the "Local Security Audit Log."

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
