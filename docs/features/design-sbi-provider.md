# Design Doc: State-Bound Identity (SBI) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Current Hardware-Locked Mission Leases (MBHL) protect the *identity* of the agent but not the *state* of the environment it operates in. An attacker can "Rug Pull" a mission by modifying project configuration or source code *after* an agent has been granted high-privilege access.

The SBI Provider anchors agency to the literal state of the workspace. By cryptographically binding identity tokens to a hash of the environment (including git HEAD and uncommitted changes), any modification to the workspace automatically invalidates active leases.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate "Workspace Hashes" representing the current state of the local environment.
    * Bind MBHL tokens to these hashes.
    * Provide sub-millisecond validation of state-integrity during tool calls.
* **Non-Goals:**
    * Version controlling the entire environment (handled by Git).
    * Preventing intentional, authorized code modifications by the agent (requires lease re-issue).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a "Write-Access" lease remains valid only as long as the underlying security rules haven't been tampered with.
* **The Happy Path (Tasks):**
    1. The agent requests a lease to run `npm deploy`.
    2. The SBI Provider generates a hash: `SHA256(git_rev + diff)`.
    3. The lease is issued, bound to this hash.
    4. An external process (or rogue subagent) modifies `.env` to redirect traffic.
    5. The agent attempts to call `npm deploy`.
    6. The SBI Provider detects a hash mismatch.
    7. The tool call is interdicted and the lease is revoked.

## 4. Design & Architecture
* **System Flow:**
    [Tool Call] -> [SBI Validator] -> [Hash Comparison] -> [Policy Engine (Allow/Deny)]
* **APIs / Interfaces:**
    * `GET /v1/auth/workspace/hash`: Returns the current environment fingerprint.
    * Internal Hook: `ValidateLease(token, current_hash)`.
* **Data Storage/State:**
    * Workspace hashes are ephemeral and session-bound, stored in kernel-locked memory.

## 5. Alternatives Considered
* **Continuous File Watching:** Rejected due to performance overhead and race conditions (TOCTOU). Hash-at-call-time is more deterministic.
* **Git-Only Pinning:** Rejected as it misses uncommitted changes (like `.env` or temporary scripts), which are common attack vectors.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The hashing logic must run in the `Deterministic Attestation Gateway` to prevent subagent tampering.
* **Observability:** Hash mismatch events are logged as "Integrity Violations" in the Local Security Audit Log.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
