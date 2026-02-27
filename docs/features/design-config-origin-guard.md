# Design Doc: Config Origin Guard (Immutable Base)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent security vulnerabilities in agentic coding tools (e.g., Claude Code CVE-2026-21852) have shown that allowing untrusted project repositories to override core configuration parameters (like API base URLs or credential paths) can lead to API key exfiltration and remote code execution. MCP Any needs a mechanism to distinguish between "System-Level" (Trusted) and "Project-Level" (Untrusted) configurations, ensuring that critical security boundaries cannot be modified by local configuration files.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement an "Immutable Base" for sensitive configuration fields.
    * Provide a clear hierarchy of configuration origins (System > User > Project).
    * Warn or block project-level overrides for "Protected Fields" (e.g., `ANTHROPIC_BASE_URL`, `OPENAI_API_BASE`, `MCPANY_VAULT_PATH`).
    * Ensure that a project-level `mcpany.yaml` can only add tools or local settings, not modify global security posture.
* **Non-Goals:**
    * Replacing the entire configuration system.
    * Preventing all local configuration (local tools are still allowed, just restricted).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer.
* **Primary Goal:** Open an untrusted open-source repository and use MCP Any without risking credential theft.
* **The Happy Path (Tasks):**
    1. User starts MCP Any in a new, untrusted repository.
    2. The repository contains a `.mcpany/config.yaml` that attempts to redirect the Anthropic API to a malicious endpoint.
    3. MCP Any detects the attempt to override a "Protected Field" from a project-level origin.
    4. MCP Any ignores the malicious override and logs a security warning to the user.
    5. The agent continues to function using the trusted system-level API base.

## 4. Design & Architecture
* **System Flow:**
    - **Origin Tracking**: Every configuration value is tagged with its origin (FILE_SYSTEM, FILE_USER, FILE_PROJECT, ENV).
    - **Protection Policy**: A hardcoded or system-level defined list of "Protected Fields" is maintained.
    - **Merge Logic**: The configuration merger skips or rejects updates to protected fields if the new value comes from a lower-trust origin (PROJECT).
* **APIs / Interfaces:**
    - Internal `ConfigManager.LoadWithProvenance()` method.
    - CLI `mcpany config origins` to inspect which values came from where.
* **Data Storage/State:** Configuration provenance is kept in-memory during the server lifecycle.

## 5. Alternatives Considered
* **Disabling Project-Level Config Entirely**: Too restrictive; users want to define project-specific tools (e.g., a "test runner" for a specific repo).
* **User Confirmation for Every Project**: High friction; "Do you trust this repository?" prompts are often ignored (security fatigue).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero-Trust component. It enforces that "Environment Trust" cannot be escalated by external content.
* **Observability:** Security violations (attempted overrides) must be logged with high priority and surfaced in the UI.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation. Triggered by Claude Code vulnerability research.
