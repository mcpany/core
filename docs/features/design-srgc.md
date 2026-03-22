# Design Doc: Semantic Reasoning Garbage Collector (SRGC)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
In deep agent swarms, the "Blackboard" (Shared KV Store) often becomes polluted with hallucinations, redundant reasoning fragments, and "Context Smears" from divergent subagent branches. This leads to token exhaustion and reasoning degradation.

SRGC is a background state management service that continuously evaluates the semantic integrity of fragments stored on the Blackboard. It identifies and prunes data that diverges from the "Mission-Root" manifest or exhibits high entropy, ensuring the swarm operates on a "clean" cognitive state.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically identify and prune "Dirty State" (hallucinations, drift) from the Blackboard.
    * Maintain alignment between subagent state and the Hardware-Attested Mission Manifest (HAMM).
    * Reduce token overhead by reclaiming context window space from redundant fragments.
* **Non-Goals:**
    * Replacing the Blackboard's primary isolation model (RAMS).
    * Modifying agent reasoning traces *before* they are committed to the Blackboard.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Engineer
* **Primary Goal:** Prevent "Context Smearing" from causing a cognitive meltdown in a long-running research swarm.
* **The Happy Path (Tasks):**
    1. A specialized subagent begins "Refinement Drift," committing high-entropy reasoning fragments to a shared shard.
    2. SRGC detects the drift by comparing fragments against the Mission-Root intent.
    3. SRGC flags the high-entropy fragments and marks them for "Garbage Collection."
    4. The Blackboard forcefully purges the marked fragments.
    5. The subagent is forced to re-align with the parent intent or is terminated by the Reaper.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Blackboard Commit] --> B{SRGC Monitor}
        B --> C[Semantic Entropy Analysis]
        C --> D[Mission-Root Alignment Check]
        D --> E{Threshold Exceeded?}
        E -- Yes --> F[Flag for Pruning]
        E -- No --> G[Commit Validated]
        F --> H[Atomic Purge]
    ```
* **APIs / Interfaces:**
    * `SRGC_Analyze_Fragment(fragment_id)`: Manual trigger for semantic analysis.
    * `SRGC_Policy_Update(config)`: Configure pruning thresholds and mission-anchors.
* **Data Storage/State:**
    * "Entropy Scores" metadata attached to Blackboard fragments.

## 5. Alternatives Considered
* **Periodic Manual Rollbacks**: Rejected as non-scalable for machine-speed swarms.
* **Aggressive Context Compression**: Complements SRGC but doesn't solve the "Dirty Data" problem, as it might compress and keep hallucinations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SRGC must be hardware-attested to prevent malicious subagents from pruning legitimate "Audit" fragments.
* **Observability:** Pruning events are logged in the "SGC Lifecycle Visualizer" (UI).

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
