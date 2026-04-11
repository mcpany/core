# Design Doc: Hierarchical Mission Anchor (HMA) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms grow in depth (parent -> subagent -> sub-subagent), maintaining the semantic integrity of the original mission becomes increasingly difficult. "Mission Drift" occurs when sub-specialists begin acting on misinterpreted intents or stale fragments, leading to unauthorized tool execution or hallucinatory loops.

The HMA Manager solves this by providing a hardware-locked, recursive lineage for every agent in the chain. By anchoring every reasoning step and tool call to a monotonic counter rooted in the user's initial mission, MCP Any ensures that no subagent can diverge from its authorized scope without triggering an immediate integrity failure.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-locked monotonic counters for all agent delegations.
    * Provide cryptographically signed lineage tokens that persist across 10+ depth hops.
    * Automate the termination of sub-intents that exhibit >20% semantic drift from the HMA.
* **Non-Goals:**
    * Will not perform real-time LLM reasoning on every instruction (delegated to ARI Hub).
    * Will not replace transport-layer security (mTLS/Named Pipes).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Code Refactor" subagent from spawning a "Cloud Exfiltrator" sub-subagent via intent injection.
* **The Happy Path (Tasks):**
    1. User initiates a mission with a TPM-signed Root Mission Anchor.
    2. Primary agent spawns a Refactor subagent; HMA Manager issues a Level-1 Nested Token.
    3. Refactor subagent attempts to spawn a sub-subagent with an unauthorized intent.
    4. HMA Manager detects the lineage mismatch via monotonic counter validation.
    5. HMA Manager interdicts the spawn request and alerts the mission-root.

## 4. Design & Architecture
* **System Flow:**
    [Root Mission] -> [HMA: 001] -> [Subagent A: HMA: 001.01] -> [Subagent B: HMA: 001.01.01]
* **APIs / Interfaces:**
    * `/v1/hma/mint`: Issue a nested lineage token.
    * `/v1/hma/verify`: Validate a tool call against its mission-root lineage.
* **Data Storage/State:**
    * Monotonic counters stored in kernel-bound secure memory.
    * Lineage fragments cached in the Blackboard (Shared KV Store) with RID-compliant isolation.

## 5. Alternatives Considered
* **Flat Identity Tokens:** Rejected because they do not track the depth or specific branch of the reasoning path, making them vulnerable to "Identity Spoofing" by specialized agents.
* **Pure Semantic Analysis:** Rejected due to the latency of LLM-based drift detection; hardware-locked counters provide sub-millisecond interdiction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage tokens are TPM-bound and non-reusable across mission branches.
* **Observability:** Integrated with the **Mesh-Resident Lineage Tracker** for real-time visualization of agent chains.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
