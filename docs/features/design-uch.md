# Design Doc: Universal Context Harmonizer (UCH)
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
As AI agent swarms evolve from linear execution to high-density horizontal meshes, maintaining a consistent mission state across heterogeneous frameworks (OpenClaw, Claude Code, AutoGen) has become a primary bottleneck. Current sharding strategies (AMS/PAMS) resolve transport-layer locks but do not address **Semantic Dissonance**—where different agents summarize or compact the same context fragment using conflicting logic.

The **Universal Context Harmonizer (UCH)** is a framework-neutral state synchronization service designed to resolve these dissonances. It leverages the new OpenClaw Cognitive Mirroring hooks to monitor subagent reasoning and ensure that compacted context fragments remain semantically aligned with the hardware-attested mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Resolve conflicting context summaries between disparate agent frameworks.
    * Provide a standardized "Harmonization API" for pluggable ContextEngines.
    * Maintain a hardware-attested "Cognitive Mainline" for the mission.
    * Reduce Mean Time to Coordinate (MTTC) by pre-emptively aligning state.
* **Non-Goals:**
    * Replacing existing framework-specific summarization logic.
    * Managing the underlying physical memory transport (handled by DME/ZCMB).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Ensure a Claude-led team and OpenClaw specialists share a unified view of a complex codebase refactor.
* **The Happy Path (Tasks):**
    1. The Mission Root initiates a multi-framework swarm.
    2. Specialist agents generate local context summaries during their task execution.
    3. The UCH intercepts these summaries via Cognitive Mirroring hooks.
    4. The UCH detects a dissonance between a Claude "Action Summary" and an OpenClaw "Schema Summary."
    5. The UCH applies the "Harmonization Policy" to merge the fragments into a consistent view.
    6. All agents receive the harmonized state update, preventing reasoning divergence.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Framework A] -->|Local Summary| B[UCH Gateway]
        C[Agent Framework B] -->|Local Summary| B
        B -->|Cognitive Mirroring Check| D[Harmonization Engine]
        E[Mission Root Manifest] -->|Alignment Policy| D
        D -->|Harmonized State| F[Shared Blackboard]
        F -->|Global Update| A
        F -->|Global Update| C
    ```
* **APIs / Interfaces:**
    * `POST /v1/harmonize`: Submits a context fragment for alignment.
    * `GET /v1/dissonance`: Streams real-time logic divergence alerts.
* **Data Storage/State:** Harmonized fragments are stored in the **Intent-Sealed Blackboard Shards**, tagged with a "Harmonization Version" for rollback.

## 5. Alternatives Considered
* **Global Locking**: Rejected due to 2s+ MTTC latency in horizontal swarms.
* **Master-Slave Summarization**: Rejected as it creates a cognitive bottleneck and ignores framework-specific specialization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All harmonization requests must carry a hardware-attested **Reasoning Path Watermark (RPW)** to prevent state injection.
* **Observability:** Dissonance events and harmonization merges are logged in the **Mission-Root Lineage Tracker** for forensic auditing.

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
