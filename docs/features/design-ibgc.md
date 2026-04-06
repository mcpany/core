# Design Doc: Intent-Bound Garbage Collection (IBGC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLM context windows scale to 1M+ tokens, aggressive garbage collection (GC) and summarization become necessary to maintain performance and reduce costs. However, standard GC often suffers from "Instruction Eviction," where critical mission-root constraints or behavioral guardrails (Silent Anchors) are pruned because they haven't been recently referenced. This leads to "GC Fragility," where agents lose their cognitive boundaries mid-session.

Intent-Bound Garbage Collection (IBGC) solves this by performing real-time semantic analysis of the context window to distinguish between transient "noise" and mission-critical "anchors."

## 2. Goals & Non-Goals
* **Goals:**
    * Perform semantic-aware pruning of the LLM context window.
    * Protect "Mission Anchors" from eviction until mission completion.
    * Reduce token consumption by 40% without degrading reasoning quality.
    * Provide "GC-Immune" pinning for user-defined safety constraints.
* **Non-Goals:**
    * Replacing the model's native context management; it acts as a pre-processing middleware.
    * Managing long-term episodic memory (handled by UMMB/UEG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Long-Running Autonomous Swarm
* **Primary Goal:** Maintain mission integrity over a 48-hour continuous coding session.
* **The Happy Path (Tasks):**
    1. User initiates a mission with specific "Hard Constraints" (e.g., "Do not delete production data").
    2. IBGC identifies these constraints and marks them as "Mission Anchors."
    3. As the session progresses and the context window fills, the IBGC Middleware scans the reasoning trace.
    4. It identifies high-entropy "exploration" noise that is no longer relevant to the current task.
    5. During the GC cycle, IBGC prunes the noise but "pins" the Mission Anchors, re-injecting them into the compact context.
    6. The agent maintains the safety constraint even after thousands of intermediate tool calls.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Raw Context] --> B[IBGC Middleware]
        B --> C{Semantic Analyzer}
        C -->|Anchor| D[Protected Buffer]
        C -->|Noise| E[Pruning Queue]
        D --> F[Compact Context Output]
        B --> G[Mission-Root Manifest]
    ```
* **APIs / Interfaces:**
    * `ibgc.RegisterAnchor(fragmentID, ttl)`: Explicitly pins a fragment.
    * `ibgc.PruneContext(currentContext, intent) -> CompactContext`: Performs the GC logic.
* **Data Storage/State:**
    * **Anchor Registry:** In-memory store of semantically significant fragments bound to the session ID.

## 5. Alternatives Considered
* **Last-Recently-Used (LRU) Pruning:** Rejected because it frequently evicts the mission root (the oldest fragment).
* **Summarization-Only:** Rejected because summarization often "smears" hard constraints into vague generalities, weakening guardrails.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pruning logic must not accidentally remove security-context headers or attestation tokens.
* **Observability:** Integrated with the "GC-Immune Anchor Visualizer" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
