# Design Doc: Pre-Trust Config-Loading Isolation (PTCLI)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
Today's AI agent frameworks (e.g., Claude Code) often automatically ingest project-local configuration files (like `.claude/settings.json`) during initialization. Recent disclosures (CVE-2026-33068) prove that this "Configuration-before-Trust" pattern allows malicious repositories to grant themselves elevated permissions before the user can even see a trust dialog.

MCP Any needs to solve this by acting as an authoritative "Configuration Quarantine." PTCLI ensures that no agent-adjacent settings are exposed to the reasoning loop or the tool execution environment until a hardware-bound, user-verified trust handshake is completed.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and sandbox all repository-local configuration loading.
    * Block ingestion of configuration hooks and automated tool sequences until explicit user attestation.
    * Provide a cryptographically signed "Safe-to-Load" manifest after user approval.
    * Align with the NVIDIA OpenShell runtime standard for policy-based guardrails.
* **Non-Goals:**
    * Replacing the agent's internal configuration parser (we wrap/proxy it).
    * Managing global user settings (PTCLI focuses on project-local overrides).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Securely clone and run an agent on a third-party repository without risk of automatic permission escalation.
* **The Happy Path (Tasks):**
    1. User points MCP Any to a new repository.
    2. MCP Any detects a project-local configuration file (e.g., `.mcpany/settings.yaml`).
    3. PTCLI immediately moves the file to a "Quarantine Shard" (isolated from the agent).
    4. PTCLI presents a visual diff of the configuration to the user via the UI.
    5. User reviews the requested permissions and hooks.
    6. User approves the configuration using a hardware-bound key (TPM/SEP).
    7. PTCLI signs the configuration and releases it to the agent's execution environment.

## 4. Design & Architecture
* **System Flow:**
    `[Project Repository] -> [PTCLI Interceptor] -> [Quarantine Shard] -> [User Approval (UI)] -> [Signed Manifest] -> [Agent Runtime]`
* **APIs / Interfaces:**
    * `POST /v1/config/quarantine`: Register a project-local config file.
    * `GET /v1/config/review/{id}`: Retrieve a diff for user review.
    * `POST /v1/config/attest`: Submit a hardware signature to release the config.
* **Data Storage/State:**
    * Configurations are stored in encrypted, isolated SQLite shards (Blackboard).
    * State is locked until the `is_attested` bit is set via the Hardware Trust Provider.

## 5. Alternatives Considered
* **Passive Warning:** Just showing a warning in the UI. Rejected because agents often load configs asynchronously, leading to TOCTOU races.
* **Static Analysis Only:** Scanning for "dangerous" keys. Rejected because "dangerous" is context-dependent and obfuscated WASM hooks can bypass simple scanners.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PTCLI follows the "Verify-before-Discovery" principle. No configuration data is mapped into the agent's discovery bus until attestation.
* **Observability:** Every quarantine and attestation event is logged in the `Local Security Audit Log` with hardware provenance.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
