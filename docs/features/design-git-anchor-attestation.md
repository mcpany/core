# Design Doc: Git-Anchor Attestation Provider
**Status:** Draft
**Created:** 2026-04-24

## 1. Context and Scope
With the rise of "Agentic Honeypots" and "Git-Diff Injection" attacks, it is no longer sufficient to secure the agent's runtime environment alone. Malicious actors can weaponize the repository itself by modifying the `.git` index or injecting subtle changes between a user's `git verify` and the agent's boot.

The Git-Anchor Attestation Provider (GAAP) cryptographically links the AI agent's session to a specific, verified state of the project's Git repository. It ensures that the agent is operating on the exact code and configuration that the user expects, providing a "Negative Trust" guarantee against environmental poisoning.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a "Git-State Fingerprint" comprising the current commit hash and a SHA-256 hash of the Git index.
    * Mandate "Git-Anchor Validation" as a prerequisite for the Deterministic Boot sequence.
    * Detect unauthorized modifications to the project environment (e.g., untracked files in sensitive directories) before agent execution.
    * Provide a signed attestation of the Git state to the Agent Runtime.
* **Non-Goals:**
    * Managing Git operations (commit, push, pull). GAAP is a read-only validation service.
    * Solving merge conflicts or handling branching logic beyond identifying the current state.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Verify that the agent is not operating on a "poisoned" index or unauthorized local modifications.
* **The Happy Path (Tasks):**
    1. The user prepares to start a Claude Code or OpenClaw session in a sensitive repository.
    2. MCP Any's GAAP middleware triggers during the Pre-Flight check.
    3. GAAP retrieves the `HEAD` commit hash and computes a hash of the current Git index (`git ls-files --stage`).
    4. GAAP compares this state against the "Last Known Good" state (if configured) or prompts the user to attest to the current Git state.
    5. Finding the state verified, GAAP generates a signed "Git-Anchor Manifest."
    6. The agent boots, with its reasoning and tool access cryptographically bound to this Git state.

## 4. Design & Architecture
* **System Flow:**
    `[Boot Trigger] -> [GAAP] -> [Git State Extraction] -> [Index Hashing] -> [Hardware Signer] -> [Signed Anchor]`
* **APIs / Interfaces:**
    * `GetGitStateFingerprint() (GitFingerprint, error)`: Internal service to extract commit and index hashes.
    * `VerifyGitAnchor(expectedFingerprint GitFingerprint) bool`: Validation gate for the boot sequence.
* **Data Storage/State:**
    * GAAP maintains a "Verified State Log" in the MCP Any database to track the evolution of project anchors.

## 5. Alternatives Considered
* **Full FS Snapshotting:** Rejected because it is too resource-intensive for large repositories. Git index hashing provides a high-fidelity signal with minimal overhead.
* **Manual User Review of Diffs:** Rejected as the primary control due to "Attestation Fatigue." GAAP automates the verification, only alerting the user on unexpected changes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If the `.git` directory itself is compromised, GAAP's signals could be spoofed. We mitigate this by including hardware-bound timestamps and signing.
* **Observability:** Any drift in Git state during an active session (e.g., external process modifying files) triggers a "State Integrity Violation" and halts the agent.

## 7. Evolutionary Changelog
* **2026-04-24:** Initial Document Creation.
