# Design Doc: Silent Anchor Guard (SAG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward deeper and more horizontal collaboration, the integrity of the "Mission Root" becomes the primary security frontier. Recent "Shadow-Root" exploits demonstrate that malicious sub-processes can redefine an agent's behavioral guardrails by injecting deceptive configuration fragments into shared teammate mailboxes or sharded context.

The Silent Anchor Guard (SAG) is designed to protect these mission-critical behavioral guardrails. By providing hardware-locked, immutable context fragments that persist across all sub-mission branches, SAG ensures that an agent's core instructions and safety policies remain sovereign, even if the surrounding context is compromised or evicted during garbage collection.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested immutability for specific "Anchor" context fragments.
    * Ensure anchors persist across recursive delegation hops and sub-process spawns.
    * Neutralize "Shadow-Root" exploits by prioritizing hardware-locked anchors over conflicting runtime configurations.
    * Prevent "Instruction Eviction" during aggressive context-window garbage collection.
* **Non-Goals:**
    * Managing non-critical runtime state or temporary agent memory.
    * Replacing existing transport-layer security (mTLS, etc.).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Ensure that a 10-agent swarm executing a complex financial audit cannot have its "Audit Integrity Policy" overridden by a compromised specialist subagent.
* **The Happy Path (Tasks):**
    1. Administrator defines a "Mission Root" with an attached "Audit Integrity Policy" marked as a Silent Anchor.
    2. MCP Any generates a hardware-attested (TPM/Secure Enclave) hash of the Anchor fragment.
    3. The primary agent delegates a task to a specialist subagent.
    4. The subagent attempts to inject a "Shadow Root" configuration that relaxes audit constraints.
    5. The Silent Anchor Guard detects the conflict and forcefully overrides the subagent's injection with the hardware-locked Anchor.
    6. The mission proceeds under the original constraints, and a "Shadow-Root Attempt" is logged in the security dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Initiation] --> B{SAG Policy Engine}
        B -->|Hardware Seal| C[TPM/Secure Enclave]
        B --> D[Active Context Window]
        D --> E[Subagent Spawn]
        E --> F{Context Merger}
        F --> G[Subagent-provided Context]
        F -->|Override Conflict| B
        F --> H[Final Reasoning Context]
    ```
* **APIs / Interfaces:**
    * `POST /v1/anchors`: Register a new immutable context anchor.
    * `GET /v1/anchors/verify`: Verify the hardware-attested integrity of current anchors.
* **Data Storage/State:**
    * Anchors are stored in kernel-bound memory and indexed by their hardware-attested SHA-256 fingerprints.

## 5. Alternatives Considered
* **Runtime Regex Filtering:** Rejected due to high latency and the risk of sophisticated "Soldered Instruction" bypasses.
* **Static Configuration Files:** Rejected because they cannot handle the dynamic nature of mission-root delegation in horizontal swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SAG utilizes hardware-root-of-trust to ensure anchors are untamperable. All context merges are subject to "Anchor Priority" rules.
* **Observability:** Every anchor verification failure or override event is logged with full lineage metadata for forensic auditing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
