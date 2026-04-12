# Design Doc: Dynamic Shard Anchoring (DSA) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In large-scale agent swarms, "Attention Drift" often occurs as the context window fills with low-utility reasoning traces from specialist subagents, leading to the eviction of critical mission-root behavioral guardrails. While static pinning (ALRA) provides a baseline, it is inefficient for long-running tasks where the "most relevant" context changes as the mission progresses through different phases.

The Dynamic Shard Anchoring (DSA) Provider evolves the attention governance layer into an active, intent-aware system. It performs real-time semantic analysis of the agent's current reasoning path and dynamically rotates relevant context shards into higher-priority attention tiers while moving dormant fragments to lower-priority "cold" storage.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically prioritize context shards based on real-time intent analysis.
    * Reduce "Cognitive Stall" by ensuring mission-critical behavioral guardrails are always present in the active attention window.
    * Integrate with the `AEM` (Agentic Entropy Monitor) to detect when intent drift requires a shard rotation.
    * Provide hardware-attested proof of shard priority transitions.
* **Non-Goals:**
    * Replacing the underlying KV-cache management (DSA orchestrates existing APIs).
    * Modifying model weights or training data.
    * Managing the content of shards (handled by `ContextEngine`).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Maintain strict security guardrails during a multi-phase "Research -> Code -> Test" mission where the relevant security policies shift from "Data Privacy" to "Safe Execution."
* **The Happy Path (Tasks):**
    1. The Architect defines multiple "Context Shards" (e.g., `policy_privacy`, `policy_execution`).
    2. The agent begins the "Research" phase; the DSA Provider identifies the intent and elevates `policy_privacy` to the P0 attention tier.
    3. As the agent transitions to "Coding," the `AEM` signals an intent shift.
    4. The DSA Provider automatically rotates `policy_execution` into the P0 tier, moving `policy_privacy` to P1.
    5. The mission completes without the agent ever losing access to the most relevant guardrails despite a 1M+ token window.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Reasoning Trace] --> B[Intent Analyzer]
        C[AEM Signal] --> B
        B --> D{DSA Provider}
        D -->|Elevate| E[P0 Attention Tier]
        D -->|Demote| F[P1 Attention Tier]
        E --> G[LLM Context Window]
        F --> G
    ```
* **APIs / Interfaces:**
    * `POST /v1/dsa/rotate`: Forces a shard rotation based on a specific intent-signal.
    * `GET /v1/dsa/status`: Returns current attention tier mappings for all shards.
* **Data Storage/State:**
    * Shard priority mappings are stored in the hardware-attested segment of the Blackboard.

## 5. Alternatives Considered
* **Static Pinning (ALRA only):** Rejected as it leads to "Attention Saturation" where the active window is filled with irrelevant but pinned fragments.
* **Frequent Context Compaction:** Rejected due to the high token cost and latency of repeated summarization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shard rotations must be verified against the Mission Manifest to prevent "Context Injection" attacks where a subagent tries to elevate its own instructions to P0.
* **Observability:** Shard rotation events are visualized in the "DSA Monitoring Dashboard."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
