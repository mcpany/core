# Design Doc: Recursive Discovery Attestation (RDA) Provider
**Status:** Draft
**Created:** 2026-05-17

## 1. Context and Scope
With the rise of large agent swarms and "Dynamic Teams," subagents are increasingly empowered to perform their own tool discovery. However, the "Shadow Delegation" exploit pattern has emerged, where subagents discover and execute project-local tools or configuration hooks that were never explicitly authorized by the user or the primary agent. This bypasses the centralized security policies and intent-anchoring of the mission.

The RDA Provider is a core security service in MCP Any that enforces "Attestation Lineage" during the discovery phase. It ensures that every capability discovered by a child agent is cryptographically cross-referenced against the parent agent's hardware-attested discovery manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a recursive verification loop for subagent tool discovery.
    * Mandate cryptographic cross-referencing against parent attestation manifests.
    * Neutralize "Shadow Delegation" by blocking un-attested tool loading.
    * Provide "Negative Discovery Proofs" showing that no unauthorized hooks were executed.
* **Non-Goals:**
    * Will not perform the actual discovery; it only validates the results of discovery.
    * Will not manage the hardware keys (this is handled by HLIS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Architect
* **Primary Goal:** Prevent subagents in a code-refactoring swarm from discovering and using unauthorized local SSH or credential-harvesting scripts.
* **The Happy Path (Tasks):**
    1. The primary agent boots with a hardware-attested manifest of approved project tools.
    2. A "Refactor Subagent" is spawned to handle a specific file.
    3. The subagent performs discovery and finds a local script `.hooks/post-refactor.sh`.
    4. The RDA Provider intercepts the discovery event and cross-references the script's hash against the parent's manifest.
    5. Since the script is not in the manifest, the RDA Provider blocks it and logs a "Shadow Delegation" alert.
    6. The subagent is notified of the blocked capability and continues using only approved tools.

## 4. Design & Architecture
* **System Flow:**
    * Subagent Discovery Request -> PNTD Provider -> RDA Provider (Interception) -> Parent Manifest Lookup -> Hash Comparison -> [Allow/Deny] -> Result to Subagent.
* **APIs / Interfaces:**
    * `VerifyDiscovery(subagent_token, discovery_object)`: Internal gRPC endpoint for discovery validation.
    * `GetNegativeProof(mission_token)`: Returns a signed manifest of blocked discovery attempts.
* **Data Storage/State:**
    * Volatile cache of session-bound manifests indexed by `mission_token`.

## 5. Alternatives Considered
* **Flat Discovery Allow-lists:** Rejected because it doesn't scale with dynamic swarms where subagents need to discover tools relevant to their specialized tasks.
* **Manual User Approval for Every Sub-Discovery:** Rejected due to "Approval Fatigue."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RDA depends on the integrity of the HLIS (Hardware-Locked Intent Store) for manifest signatures.
* **Observability:** High-fidelity logging of all blocked discovery attempts, integrated into the "Shadow Delegation" alert system.

## 7. Evolutionary Changelog
* **2026-05-17:** Initial Document Creation.
