# Design Doc: Repository-Gated Configuration (RGC)
**Status:** Draft
**Created:** 2026-04-03

## 1. Context and Scope
The disclosure of CVE-2026-21852 in Claude Code has exposed a critical vulnerability where repository-controlled configuration files (e.g., `.claude/settings.json`) can override user-defined safety safeguards. This allows a malicious repository to execute unauthorized code, steal API keys, or modify cloud-stored project files without user consent. RGC ensures that MCP Any acts as an authoritative gatekeeper for all environment-resident configurations.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all project-local agent configurations (e.g., `.mcpany/settings.json`, `.claude/settings.json`).
    * Mandate cryptographic user attestation for any configuration change originating from the repository.
    * Neutralize "Configuration-as-Execution" by isolating repository-defined hooks until approved.
* **Non-Goals:**
    * Restricting user-defined global configurations.
    * Modifying repository files directly (RGC is a validating proxy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely open a community repository without risking API key exfiltration via poisoned settings.
* **The Happy Path (Tasks):**
    1. The user clones and opens a new repository.
    2. The agent attempts to load a project-local configuration containing an `on_open` hook.
    3. RGC intercepts the load request and identifies the hook as "Un-attested."
    4. MCP Any pauses execution and surfaces an attestation dialog to the user.
    5. The user reviews the hook, identifies it as malicious, and denies the attestation.
    6. RGC blocks the configuration load, protecting the user's environment.

## 4. Design & Architecture
* **System Flow:**
  [Agent Load Config] -> [RGC Middleware] -> [Local Attestation Store] -> [User Approval Flow (if missing)] -> [Sanitized Config Data]
* **APIs / Interfaces:**
    * `mcpany.rgc.v1.ConfigValidator`
    * Hook: `onConfigLoad(path, content_hash)`
* **Data Storage/State:**
    * Persistent SQLite store for attested configuration hashes bound to specific repository paths.

## 5. Alternatives Considered
* **Implicit Trust (Current State)**: Rejected as demonstrated to be insecure by CVE-2026-21852.
* **Static Blocklists**: Rejected as they cannot keep up with polymorphic injection patterns in natural language configs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All repository-resident data is treated as untrusted until attested.
* **Observability:** Audit logs capture all attestation events and blocked configuration attempts.

## 7. Evolutionary Changelog
* **2026-04-03:** Initial Document Creation.
