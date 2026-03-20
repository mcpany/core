# Design Doc: Reason-Graph Integrity (RGI) Provider
**Status:** Draft
**Created:** [2026-06-18]

## 1. Context and Scope
The emergence of horizontal teammate coordination in frameworks like OpenClaw and Claude Code has introduced the concept of "Reason-Graphs"-- distributed, parallel reasoning paths where multiple specialists contribute to a shared objective. However, this has led to **Reason-Graph Collision (RGC)**, where parallel teammates with overlapping roles generate conflicting reasoning traces that cannot be reconciled by simple binary state handoffs.

The Reason-Graph Integrity (RGI) Provider acts as the authoritative "Graph Arbiter" for MCP Any. It provides the infrastructure to merge parallel reasoning traces, resolve semantic conflicts at the graph level, and ensure that the collective swarm intelligence remains anchored to the mission-root intent without cognitive stall.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested graph-conflict resolution strategies for parallel teammates.
    * Provide a standardized interface for merging non-conflicting reasoning traces (CFRR-compatible).
    * Detect and resolve Reason-Graph Collisions (RGC) before they lead to cognitive stall.
    * Ensure mission-root sovereignty across complex, multi-threaded reasoning paths.
* **Non-Goals:**
    * Managing low-level transport (handled by Isolated Named Pipes).
    * Providing individual agent memory (handled by ContextEngine).
    * Performing real-time entropy gating (handled by AAG Middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Orchestrator
* **Primary Goal:** Successfully merge conflicting reasoning paths from two specialists (e.g., a "Security Auditor" and a "Performance Engineer") without losing mission context.
* **The Happy Path (Tasks):**
    1. Mission Root spawns two parallel teammates: Agent S (Security) and Agent P (Performance).
    2. Both agents contribute to the Reason-Graph on the Shared Blackboard.
    3. Agent S proposes a restrictive security patch; Agent P proposes a conflicting performance optimization for the same block.
    4. RGI Provider detects the collision during the graph-sync phase.
    5. RGI utilizes the mission-root priority policy to arbitrate the conflict.
    6. RGI merges the non-conflicting fragments and provides a "Winning Graph" to the swarm.
    7. Coordination continues without manual intervention or cognitive stall.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
        graph TD
        A[Teammate A Trace] --> B[RGI Hub]
        C[Teammate B Trace] --> B[RGI Hub]
        B --> D[Conflict Detector]
        D --> E{Collision?}
        E -- Yes --> F[Mission-Root Policy Arbiter]
        E -- No --> G[Atomic Graph Merge]
        F --> H[Resolved Reasoning Fragment]
        H --> G
        G --> I[Unified Reason-Graph]
        I --> J[Hardware-Attested Snapshot]
    ```
* **APIs / Interfaces:**
    * `rgi.ProposeFragment(graphId, fragment) -> void`: Adds a reasoning fragment to the active mesh graph.
    * `rgi.SynchronizeGraph(graphId) -> UnifiedGraph`: Triggers collision detection and graph merging.
* **Data Storage/State:**
    * **Reason-Graph Cache:** A transient, graph-based representation of the active teammate coordination, stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Last-Writer-Wins (LWW):** Rejected because it causes loss of critical reasoning context in high-stakes missions.
* **Sequential Hand-offs:** Rejected as it eliminates the performance benefits of horizontal parallel coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All graph-merging decisions must be signed by the hardware-attested RGI Provider (MRA-compliant).
* **Observability:** Integrated with the "Reason-Graph Visualizer" for real-time auditing of conflict resolution events.

## 7. Evolutionary Changelog
* **[2026-06-18]:** Initial Document Creation. Introducing Reason-Graph Integrity (RGI) to resolve coordination collisions in horizontal meshes.
