# Design Doc: Attention-Locked Context Sharding (ALCS)
**Status:** Draft
**Created:** 2026-06-15

## 1. Context and Scope
The persistent risk of **Reasoning Entropy Exhaustion (REE)** and "Context-Window Flooding" has proven that standard context isolation is insufficient for deep swarms. As subagents generate high-entropy reasoning traces, they can inadvertently or maliciously evict "Mission Root" intent anchors from the LLM's attention window.

Attention-Locked Context Sharding (ALCS) utilizes hardware-bound attention-locking headers to "pin" critical context shards (Mission Root, Sovereignty Proofs) to a protected "Attention Tier." This ensures that the primary mission intent remains the sovereign driver of agent reasoning regardless of subagent noise.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound (HAAL) attention-locking for Mission Root fragments.
    * Provide a "Protected Shard" API for high-priority context fragments.
    * Detect and mitigate REE-driven "Attention Erosion" in deep swarms.
    * Integrate with the Live Context Sharding middleware for granular state management.
* **Non-Goals:**
    * Increasing the physical size of the context window.
    * Managing tool discovery (handled by SDP/SMS).
    * Enforcing execution-time tool policies (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Orchestrator
* **Primary Goal:** Maintain mission-root consistency while delegating to 10+ parallel specialist subagents.
* **The Happy Path (Tasks):**
    1. Root agent initiates a mission and marks the "Root Intent" as an `Attention-Locked` shard.
    2. ALCS middleware wraps the fragment with HAAL-compliant hardware headers.
    3. Multiple specialist subagents generate high-volume reasoning traces (high entropy).
    4. The LLM's context window begins to fill; standard eviction logic triggers.
    5. The ALCS-protected shard remains "pinned" at the attention layer due to the hardware-bound lock.
    6. The agent continues to reason with the Root Intent as the primary anchor, preventing "Mission Drift."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Fragment] --> B[ALCS Shard Manager]
        B --> C[HAAL Header Generator]
        C --> D[Hardware Signer - TPM]
        D --> E[Attention-Locked Shard]
        E --> F[Context Window Integration]
        G[Subagent Reasoning Noise] --> F
        F --> H{Eviction Trigger?}
        H -- Noise --> I[Evict Unlocked Fragments]
        H -- Root Intent --> J[Retain Locked Shard]
    ```
* **APIs / Interfaces:**
    * `alcs.LockShard(fragment, priority) -> LockedShard`: Pins a fragment to the attention tier.
    * `alcs.GetAttentionUtilization() -> Metrics`: Reports the real-time pressure on the attention window.
* **Data Storage/State:**
    * **Active Shard Registry:** Tracking all currently locked fragments and their hardware-attested signatures.

## 5. Alternatives Considered
* **Recursive Intent Re-injection:** Rejected as it increases token consumption and can be "smeared" by high-entropy noise.
* **Static Context Pinning:** Rejected as it lacks the hardware-bound proof required for Zero-Trust environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention-locking must be hardware-attested to prevent subagent "Shadow-Locking" of malicious intents.
* **Observability:** Integrated with the "Context Attention Monitor" for real-time visualization of pinned fragments and noise levels.

## 7. Evolutionary Changelog
* **2026-06-15:** Initial Document Creation. Addressing REE and Attention Erosion via hardware-bound context sharding.
