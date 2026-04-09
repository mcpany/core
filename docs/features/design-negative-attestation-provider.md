# Design Doc: Negative Attestation Provider
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The "Absence-as-Exploit" vector (highlighted by CVE-2026-25725) occurs when a system's security depends on the non-existence of a file (e.g., a configuration override). If an attacker can create that file within a writable directory before the security boundary is fully enforced, they can bypass protections.

MCP Any needs to solve this because it acts as the authoritative gateway between untrusted agents and the host environment. The Negative Attestation Provider (NAP) ensures that restricted paths remain empty, preventing malicious code from injecting persistent hooks that execute with host privileges.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographically signed "Negative Manifest" proving the absence of files in restricted paths.
    * Support hardware-bound attestation (TPM/SEP) for the non-existence proofs.
    * Enable high-frequency re-validation of path absence during the agent lifecycle.
    * Standardize the "Non-Existence Proof" format for cross-framework interoperability.
* **Non-Goals:**
    * Validating the content of existing files (handled by the Deterministic Attestation Gateway).
    * Preventing all filesystem writes (NAP only proves absence at specific check-points).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a newly cloned repository has not injected a malicious `.claude/settings.json` file that bypasses the sandbox.
* **The Happy Path (Tasks):**
    1. The orchestrator triggers a "Pre-Flight" check for a workspace.
    2. NAP scans the designated restricted paths (e.g., `.claude/`, `.env/`).
    3. NAP identifies that no unauthorized configuration files exist in these paths.
    4. NAP generates a TPM-signed "Negative Attestation Receipt" listing the checked paths and their empty status.
    5. The orchestrator verifies the receipt before launching the agent.
    6. The agent boots with the confidence that no hidden hooks were injected.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Initiator[Orchestrator] -->|Query Paths| NAP[Negative Attestation Provider]
        NAP -->|Verify Absence| FS[(Filesystem)]
        NAP -->|Sign Result| TPM[Hardware Root of Trust]
        TPM -->|Signed Proof| Initiator
    ```
* **APIs / Interfaces:**
    * `GenerateNegativeProof(paths []string) (SignedProof, error)`: Checks a list of paths and returns a signed proof that they do not exist.
    * `StreamAbsenceHeartbeat(paths []string)`: Provides a real-time stream of absence confirmations using low-overhead kernel watches (e.g., inotify/fsnotify).
* **Data Storage/State:** Maintains a session-bound cache of attested path statuses in secure memory to reduce hardware signature overhead.

## 5. Alternatives Considered
* **Read-Only Mounting:** Rejected as the primary solution because many frameworks require writable workspace roots for legitimate operations (e.g., building code).
* **Baseline Snapshotted Environments:** Effective but high-overhead for local development where project directories are highly dynamic and large.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The NAP must operate with higher privilege than the agent (e.g., in a separate process or container) to ensure it cannot be blinded or manipulated by subagent filesystem actions.
* **Observability:** Failed absence checks trigger immediate "High" severity security alerts in the Connectivity & Security Dashboard and are logged to the Hardware-Attested audit trail.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
