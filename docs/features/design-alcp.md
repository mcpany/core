# Design Doc: Attention-Locked Context Pinning (ALCP)
**Status:** Draft
**Created:** 2026-06-23

## 1. Context and Scope
The introduction of 1M+ token context windows has created a new stability risk: **Attention Eviction**. During high-intensity reasoning cycles, especially when processing un-attested project data (e.g., recursive directory listings), model attention mechanisms often prune "distal" context fragments to prioritize local task data. This frequently results in the eviction of security anchors and "Safe-by-Default" system instructions.

ALCP provides "Attention Sovereignty" by utilizing hardware-bound attention-locking headers. It allows the infrastructure to cryptographically "lock" specific instruction fragments in the LLM's active attention layer, ensuring they are never pruned or ignored during context-compaction or high-effort reasoning.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested "Attention Locking" for critical context fragments.
    * Provide an interface for developers to tag fragments as `Attention-Locked`.
    * Interdict reasoning loops that attempt to "bypass" locked anchors via noise injection.
    * Maintain lock integrity across mirrored reasoning sessions.
* **Non-Goals:**
    * Managing the entire 1M+ token context window.
    * Replacing existing context compression algorithms (e.g., summarizing non-locked data).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars, ensuring security anchors are never pruned.
* **The Happy Path (Tasks):**
    1. Orchestrator defines a "Safe-by-Default" anchor containing environment isolation policies.
    2. MCP Any tags this fragment with an `x-mcp-attention-lock` header.
    3. Swarm initiates a "Context Mirror" across 3 specialist agents.
    4. During high-intensity reasoning, the model's native compaction logic attempts to prune the "distal" isolation policy.
    5. ALCP middleware detects the eviction attempt and forcefully re-injects the fragment into the high-priority attention tier.
    6. All 3 agents achieve the goal while remaining bound by the isolation policy.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Context Registry] -->|Tag Fragment| B[ALCP Manager]
        B -->|Inject Lock Header| C[LLM Transport]
        C -->|Reasoning Trace| D{Attention Monitor}
        D -->|Eviction Warning| E[Dynamic Re-injection]
        D -->|Lock Verified| F[Continue Reasoning]
    ```
* **APIs / Interfaces:**
    * `ALCP_Lock_Fragment(id, priority_tier)`: Pins a fragment to a specific hardware-attested attention tier.
    * `ALCP_Verify_Attention_Map()`: Validates that locked fragments are present in the model's current attention weights.
* **Data Storage/State:**
    * "Locked Anchor Manifest" stored in the hardware-bound Mission Manifest (HAMM).

## 5. Alternatives Considered
* **Repetitive Prompting**: Rejected because it consumes excessive tokens and doesn't guarantee the model won't eventually "tune out" the repeated text during high-effort reasoning.
* **Logit Bias Injection**: Rejected because it affects generation but not the model's *attention* to instructions, which is the root of the eviction problem.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALCP anchors are cryptographically linked to the mission-root identity. If a subagent attempts to spoof a lock, the signature check fails.
* **Observability:** Eviction attempts are logged as "Attention Density Alarms" in the Visual Attention Dashboard.

## 7. Evolutionary Changelog
* **2026-06-23:** Initial Document Creation.
