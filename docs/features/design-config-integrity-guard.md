# Design Doc: Config Integrity Guard (Double-Lock)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Recent vulnerabilities in Claude Code (CVE-2026-21852) showed that agents can be tricked into exfiltrating data or changing their behavior by malicious configuration files placed in untrusted directories. MCP Any needs a way to separate "Global Security Intent" from "Project-Specific Tasking."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Double-Lock" configuration hierarchy.
    * Prevent project-local configs (`.mcpany/config.yaml`) from overriding security-critical settings (e.g., allowed domains, sensitive env vars, transport protocols).
    * Require explicit user "trust" for new project-local configurations.
* **Non-Goals:**
    * Blocking all project-local configurations (which are useful for task-specific toolsets).
    * Implementing a full RBAC system for config files (simple "Trusted/Untrusted" is the goal).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer.
* **Primary Goal:** Open a downloaded repository and run an agent without worrying that the repo's config has redirected my API traffic to an attacker.
* **The Happy Path (Tasks):**
    1. User clones a repo and runs `mcpany`.
    2. MCP Any detects a local `.mcpany/config.yaml`.
    3. Server checks if this config hash is in the `~/.mcpany/trusted_configs` list.
    4. If not, it prompts the user: "Untrusted config detected. Allow?"
    5. Even if allowed, the server ignores any attempts in the local config to override `security_policy` or `core_transports`.

## 4. Design & Architecture
* **System Flow:**
    `[Load Global Config] -> [Identify Local Config] -> [Validate Trust] -> [Merge (Restricted)] -> [Active Config]`
* **APIs / Interfaces:**
    * CLI command: `mcpany trust <path>`
    * Config Schema: Mark specific fields as `Immutable` in the schema validator.
* **Data Storage/State:**
    * `~/.mcpany/trusted_configs` (JSON list of file hashes).

## 5. Alternatives Considered
* **Disabling local configs entirely:** Too restrictive; users want to define project-specific tools.
* **Automated scanning for "malicious" settings:** Hard to maintain and easy to bypass; the "Double-Lock" (Immutable globals) + "User Trust" (Opt-in) is more robust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Ensures that the "Root of Trust" always resides in the user's global configuration, which is much harder to hijack via a simple `git clone`.
* **Observability:** Audit logs will record when a local config attempt to override an immutable field.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
