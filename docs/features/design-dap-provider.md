# Design Doc: Deterministic Absence Proof (DAP) Provider
**Status:** Draft
**Created:** 2026-04-22

## 1. Context and Scope
The "Absence-as-Exploit" pattern (CVE-2026-25725) revealed a critical vulnerability where AI agents could inject malicious configurations by creating files that were expected to be absent at startup. Traditional security focuses on what is present; "Negative Trust" requires proving what is *not* present.

The Deterministic Absence Proof (DAP) Provider in MCP Any will generate cryptographic proofs of non-existence for restricted project-local files and directories. This ensures that an agent sandbox is clean of unauthorized "hook" files (like `.claude/settings.json` or `.clinerules`) before execution begins.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate signed "Non-Existence Manifests" for a configurable list of restricted paths.
    * Provide a mandatory "Negative Attestation" signal as part of the Deterministic Boot sequence.
    * Support recursive absence checking for restricted directories.
    * Integrate with the local TPM/Secure Enclave for hardware-bound signature generation.
* **Non-Goals:**
    * Managing the actual deletion of files (DAP is a provider of *proof*, not a cleanup tool; cleanup should be handled by the Pre-Flight Sandbox Validator).
    * Providing proofs for files outside the designated project root.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Operator
* **Primary Goal:** Ensure an agent cannot be "rug-pulled" by a malicious repository that injects configuration hooks upon cloning.
* **The Happy Path (Tasks):**
    1. The user initiates an agent session within a newly cloned repository.
    2. MCP Any's Deterministic Boot sequence triggers the DAP Provider.
    3. The DAP Provider scans the project root for a list of "Forbidden Hooks" (e.g., `.claude/settings.json`).
    4. Finding no such files, the DAP Provider generates a SHA-256 manifest of the checked (absent) paths.
    5. The manifest is signed using the host's hardware identity.
    6. The signed "Non-Existence Proof" is handed to the Agent Runtime as a prerequisite for boot.

## 4. Design & Architecture
* **System Flow:**
    `[Boot Trigger] -> [DAP Provider] -> [Filesystem Scan (Negative)] -> [Hardware Signer] -> [Signed Manifest]`
* **APIs / Interfaces:**
    * `GenerateAbsenceProof(paths []string) (SignedManifest, error)`: Core internal API.
    * `VerifyAbsenceProof(manifest SignedManifest) error`: Validation API for the runtime.
* **Data Storage/State:**
    * DAP manifests are ephemeral and session-bound. Fingerprints may be logged for audit trails.

## 5. Alternatives Considered
* **Directory Locking:** Rejected because it doesn't prevent file creation in new subdirectories and is OS-dependent.
* **Simple File Watchers:** Rejected because they are reactive (TOCTOU risk); DAP is proactive and part of the boot gate.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The list of "Forbidden Hooks" must be protected by the Global Policy Bus. If the list itself is compromised, the DAP becomes useless.
* **Observability:** Failed absence checks (finding a forbidden file) trigger a "High-Severity Security Alert" in the UI.

## 7. Evolutionary Changelog
* **2026-04-22:** Initial Document Creation.
* **2026-04-26:** Update: Hardening against Ambient Context Pollution.
    * **Context:** Market sync identified that subagents in shared swarms are prone to "Ambient Pollution" from unrelated config files.
    * **Architecture Adjustment:** DAP Generator now supports "Scope-Pinning," where a DAP manifest can be cryptographically bound to a specific Mission Intent, preventing its reuse for unrelated agent boots.
