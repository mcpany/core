# Design Doc: Settings Injection Guard
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
Recent findings in the agent ecosystem (e.g., CVE-2026-25725) have highlighted a critical vulnerability where agents automatically ingest "hooks" or "auto-execute" commands from project-local configuration files (like `.claude/settings.json`). This allows malicious repositories to execute arbitrary code on a collaborator's machine. The Settings Injection Guard is an active interception and validation layer designed to neutralize these "Rug Pull" attacks by ensuring that all project-local configurations match an attested baseline before they are exposed to the agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all read/write attempts to sensitive project-local configuration files.
    * Validate configuration blocks against a user-approved cryptographic baseline.
    * Provide a real-time notification and approval flow for new or modified settings.
    * Support "Non-Existence Proofs" to prevent the injection of malicious files into empty directories.
* **Non-Goals:**
    * Replacing the agent's internal configuration logic.
    * Validating the non-executable content (e.g., natural language instructions) within the settings.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a local LLM agent on an untrusted repository.
* **Primary Goal:** Prevent the agent from executing a malicious hook injected into the project's settings.
* **The Happy Path (Tasks):**
    1. The developer opens a new repository and starts an AI agent.
    2. The agent attempts to read `.claude/settings.json`.
    3. The Settings Injection Guard intercepts the request and calculates the hash of the file.
    4. If the hash is unknown, the guard suspends the request and prompts the developer for approval.
    5. The developer reviews the proposed settings (e.g., a custom `test` hook) and approves them.
    6. The guard signs the attestation and allows the agent to read the file.
    7. Subsequent reads within the session are allowed without interruption unless the file changes.

## 4. Design & Architecture
* **System Flow:**
    `Agent Runtime` -> `MCP Any Virtual Filesystem` -> `Settings Injection Guard` -> `Host Filesystem`
* **APIs / Interfaces:**
    * `intercept(path string, data []byte) (allowed bool, error)`: Core interception hook for configuration files.
    * `attest(hash string, policy Policy) error`: Interface for recording user approvals.
* **Data Storage/State:**
    * `attestations.db`: SQLite database for storing approved configuration hashes.
    * Integration with the `Deterministic Attestation Gateway` for Non-Existence Proofs.

## 5. Alternatives Considered
* **Static Analysis Only**: Rejected because it cannot prevent TOCTOU (Time-of-Check to Time-of-Use) attacks where the file is modified during execution.
* **OS-Level Write Protection**: Too restrictive; prevents legitimate agents from updating their own settings (e.g., history, session tokens).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All configuration files are treated as untrusted until they match a signed manifest.
* **Observability:** Integrated with the "Settings Integrity Monitor" in the UI to provide real-time alerts for injection attempts.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
