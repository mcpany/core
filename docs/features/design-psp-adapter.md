# Design Doc: Predictive State Purging (PSP) Adapter
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As speculative reasoning branches become deeper, the performance overhead of maintaining redundant fragments ('Attention-Window Flooding') has become a bottleneck for agentic speed. Simultaneously, the discovery of 'Attention-Splicing' proves that redundant speculative fragments can be weaponized to evict mission-critical anchors.

The Predictive State Purging (PSP) Adapter provides the infrastructure for proactively pruning speculative context shards based on a real-time 'Mission Utility' score before they reach the attention window.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-performance adapter for OpenClaw-compatible PSP strategies.
    * Provide real-time 'Mission Utility' scoring for all speculative context fragments.
    * Proactively prune low-utility fragments to prevent Attention-Window Flooding.
    * Integrate with the ABG service to ensure that mission-root anchors are never candidates for purging.
* **Non-Goals:**
    * Directly managing LLM inference (handled by providers).
    * Enforcing low-level transport security (handled by the Named-Pipe/WebSocket layer).
    * Sanitizing binary state (handled by the WASM-BSH Sanitizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Optimize token consumption and protect the attention window by proactively pruning redundant speculative fragments.
* **The Happy Path (Tasks):**
    1. Agent generates multiple speculative reasoning paths (branches).
    2. PSP Adapter receives the speculative fragments and their associated intent tokens.
    3. PSP Adapter calculates a 'Mission Utility' score for each fragment based on alignment with the mission root.
    4. Fragments with scores below the 'Utility Quorum' threshold are proactively purged from the high-speed buffer.
    5. Only high-utility fragments are promoted to the LLM context window.
    6. Attention-window health is maintained while token consumption is optimized.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Speculative Fragments] --> B[PSP Adapter]
        B --> C[Utility Scoring Engine]
        C --> D[Mission-Root Alignment Check]
        D --> E{Above Utility Threshold?}
        E -- Yes --> F[Promote to Context Window]
        E -- No --> G[Proactively Purge Fragment]
        H[Mission-Root Intent] --> D
    ```
* **APIs / Interfaces:**
    * `psp.ScoreFragment(fragment, missionToken) -> float`: Calculates the utility score.
    * `psp.PruneStaleBranches(missionToken) -> int`: Forcefully purges low-utility fragments.
* **Data Storage/State:**
    * **Utility Score Cache:** A session-bound, in-memory cache of calculated utility scores for active reasoning branches.

## 5. Alternatives Considered
    * **Reactive Garbage Collection (R-GC):** Rejected as it only clears fragments after they have already been ingested, failing to prevent AWF or Attention-Splicing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The PSP Adapter must be integrated with the ABG service to ensure that hardware-locked mission-root anchors are immutable to the purging engine.
* **Observability:** Integrated with the 'Cognitive Load Shedding Controller' for real-time monitoring of token optimization efficiency.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
