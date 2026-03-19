# Design Doc: Mesh-Resident Cognitive Load Balancer (MCLB)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In horizontal "Agent Teams," specialized teammates often exhibit highly variable reasoning intensities (ARE). Without active management, a single specialist can become a bottleneck while others remain idle, leading to "Cognitive Stall" and mesh-wide latency spikes. The MCLB is designed to act as a real-time scheduler that dynamically redistributes reasoning tasks across the mesh based on available cognitive capacity.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor real-time ARE (Advanced Reasoning Effort) scores across all mesh teammates.
    * Dynamically redistribute pending tasks to idle or low-load specialists.
    * Minimize mesh-wide MTTC (Mean Time To Completion) by eliminating cognitive bottlenecks.
    * Provide a unified view of mesh-level reasoning utilization.
* **Non-Goals:**
    * Modifying the internal reasoning logic of agents.
    * Managing the absolute token budgets (handled by RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Successfully complete a complex coding task using 5 teammates without any single agent stalling the entire loop.
* **The Happy Path (Tasks):**
    1. The swarm orchestrator registers a team of teammates with the MCLB.
    2. Teammates report their current reasoning intensity and queue depth to the MCLB.
    3. The MCLB identifies Teammate A is stuck in a high-entropy "Hallucination Loop" (High ARE).
    4. The MCLB identifies Teammate B is idle and has similar capability cards.
    5. The MCLB signals the orchestrator to migrate the mission-root fragment from A to B.
    6. Teammate B resumes the reasoning path, resolving the bottleneck.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        T1[Teammate 1] -->|ARE Heartbeat| MCLB[Cognitive Load Balancer]
        T2[Teammate 2] -->|ARE Heartbeat| MCLB
        T3[Teammate 3] -->|ARE Heartbeat| MCLB
        MCLB -->|Capacity Analysis| Sched[Scheduler]
        Task[New Mission Fragment] --> Sched
        Sched -->|Allocate| T2
    ```
* **APIs / Interfaces:**
    * `POST /v1/mclb/heartbeat`: Report real-time reasoning effort and queue state.
    * `GET /v1/mclb/capacity`: Query the mesh for the best-fit teammate for a new task.
    * `POST /v1/mclb/task/assign`: Bind a mission fragment to a specific teammate.
* **Data Storage/State:**
    * Teammate state and ARE metrics are stored in an in-memory, high-speed time-series buffer.

## 5. Alternatives Considered
* **Static Round-Robin:** Rejected as it does not account for the extreme variability of reasoning effort.
* **Orchestrator-Led Balancing:** Rejected as it adds too much latency and complexity to the primary reasoning loop.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Load balancing signals are only accepted from hardware-attested teammates within the same mission scope.
* **Observability:** Real-time Gantt-style charts of reasoning utilization in the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
