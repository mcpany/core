# Design Doc: Deterministic Boot Manifest Provider
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
Traditional AI agent boot sequences are vulnerable to "Shadow-Sandbox" escapes where malicious configurations are injected into the environment before the agent starts. To counter this, we need a "Deterministic Boot" sequence that proves the absolute integrity of the project environment. This provider will generate a cryptographically signed manifest of the environment's state, including proofs of non-existence for restricted files, before allowing any agent execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a signed manifest of all project-local files and their hashes.
    * Provide "Deterministic Absence Proofs" (DAP) for restricted paths (e.g., `.claude/settings.json`).
    * Block agent startup if the environment does not match the attested baseline.
* **Non-Goals:**
    * Monitoring the environment after boot (this is handled by the Resident Integrity Monitor).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure a repository hasn't been "poisoned" with malicious settings before starting an autonomous agent.
* **The Happy Path (Tasks):**
    1. User initiates `mcp-any start`.
    2. The `Deterministic Boot Manifest Provider` scans the current directory.
    3. It generates a DAP proving no malicious `.claude/settings.json` exists.
    4. A signed manifest is generated and passed to the agent as a "Trust Anchor."
    5. The agent starts, knowing its environment is clean.

## 4. Design & Architecture
* **System Flow:**
    * `Environment Scanner` -> `DAP Generator` -> `Manifest Signer (TPM)` -> `Agent Loader`
* **APIs / Interfaces:**
    * Internal Provider Interface: `GenerateManifest(path string) (*Manifest, error)`
* **Data Storage/State:**
    * Manifests are ephemeral and bound to the session.

## 5. Alternatives Considered
* **Path Allow-listing:** Rejected as it doesn't handle the "Absence-as-Exploit" vector where agents inject hooks into missing files.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Manifests must be signed by a hardware-bound key (TPM).
* **Observability:** Boot failures due to manifest mismatch are high-priority alerts.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
