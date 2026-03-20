# Design Doc: Anti-Pinning Context Guard
**Status:** Draft
**Created:** 2026-04-30

## 1. Context and Scope
The disclosure of CVE-2026-51002 (Context Pinning) revealed a critical flaw in how agents manage local state. Malicious subagents can inject instructions into the parent context that survive session resets, creating a persistent "Reasoning Backdoor." MCP Any needs a dedicated guard to enforce absolute context isolation and verifiable resets between missions.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Context Reset Verifier" that scans the active context window for persistent fragments.
    * Enforce cryptographic boundaries between mission-specific context segments.
    * Provide a signed "Clean Room Attestation" before any new agent mission begins.
    * Detect and purge "Pinned Instructions" that lack a valid mission lineage.
* **Non-Goals:**
    * Managing the agent's internal short-term memory (handled by the agent runtime).
    * Replacing the Shared KV Store (will coordinate with its isolation layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Security Swarm Operator
* **Primary Goal:** Start a new mission with a multi-agent swarm without risk of context contamination from a previous, potentially compromised session.
* **The Happy Path (Tasks):**
    1. Operator triggers a "Mission Reset" via MCP Any.
    2. The Anti-Pinning Guard executes a recursive purge of all context shards and Blackboard keys for the previous `mission_id`.
    3. The Guard performs a "Verification Scan" of the project environment (e.g., `.claude/settings.json`, `.mcpany/context/`) to ensure no persistent hooks remain.
    4. The Guard generates a signed "Clean Room Manifest."
    5. The new mission starts only after the manifest is verified by the Gateway.

## 4. Design & Architecture
* **System Flow:**
    `[Mission End] -> [Recursive Purge] -> [Environment Verification] -> [Clean Room Attestation] -> [New Mission Start]`
* **APIs / Interfaces:**
    * `Guard.VerifyReset()`: Performs the verification scan.
    * `Guard.Attest()`: Issues the signed manifest.
* **Data Storage/State:**
    * Persistence maps are stored in the TPM-backed configuration boot layer to ensure the Guard itself cannot be subverted.

## 5. Alternatives Considered
* **Virtual Machine Isolation:** Rejected due to high overhead for lightweight agent spawning.
* **Path-Only Isolation:** Rejected because context can be pinned in multimodal metadata or hidden configuration blocks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Guard is a mandatory prerequisite for the Gateway's "Mission Handshake."
* **Observability:** Failed reset attempts and "Stray Fragment" discoveries are logged to the High-Security Audit Log.

## 7. Evolutionary Changelog
* **2026-04-30:** Initial Document Creation.
