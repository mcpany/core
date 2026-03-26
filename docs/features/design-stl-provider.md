# Design Doc: Shared Task-List (STL) Provider
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
As agents transition from single-session linear execution to horizontal teammate meshes (e.g., Claude Code Agent Teams), the primary coordination bottleneck is the "Task List." Existing models rely on monolithic, framework-specific task stores that prevent cross-framework collaboration.

The STL Provider is a universal, hardware-attested coordination service that hosts the "Shared Task List" for a mission. It allows a Claude team lead to assign a task to an OpenClaw specialist, ensuring that task transitions, status updates, and state handoffs are synchronized across frameworks with absolute integrity.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a framework-neutral API for task creation, claiming, and completion.
    *   Ensure "Task-Claim Integrity" using hardware-attested session tokens.
    *   Maintain a cryptographically signed audit trail of task ownership.
    *   Support lock-free synchronization using CRDTs for high-density swarms.
*   **Non-Goals:**
    *   Implementing task reasoning (the agents reason about tasks; STL only coordinates the *state* of the tasks).
    *   Replacing the Blackboard (STL handles task metadata; Blackboard handles task data).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Delegate a security audit task from a Claude lead to an OpenClaw specialist and ensure the result is correctly synthesized back.
*   **The Happy Path (Tasks):**
    1.  Claude Team Lead creates a mission and pushes 3 tasks to the STL Provider.
    2.  OpenClaw Specialist queries the STL for "Audit" tasks.
    3.  OpenClaw Specialist claims Task #2 using a hardware-attested mission token.
    4.  STL Provider locks Task #2 and broadcasts the ownership change to the mesh.
    5.  OpenClaw Specialist completes the audit and pushes the "Success" signal and BSH pointer to STL.
    6.  Claude Team Lead is notified of completion and synthesizes the result from the Blackboard.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        A[Claude Lead] -->|Create Task| B(STL Provider)
        C[OpenClaw Specialist] -->|Claim Task| B
        B -->|Sync State| D[CRDT Task Shard]
        D -->|Update UI| E[Teammate Task List Viewer]
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/stl/tasks`: Create new tasks in the mission scope.
    *   `PATCH /v1/stl/tasks/{id}/claim`: Hardware-attested task claiming.
    *   `GET /v1/stl/tasks`: Stream task updates via WebSocket.
*   **Data Storage/State:**
    *   Backed by the Shared KV Store with a CRDT-based state reconciliation layer.

## 5. Alternatives Considered
*   **Centralized SQL Task Store:** Rejected because horizontal Agent Teams require non-blocking, distributed performance during "Network Partition" scenarios.
*   **Direct A2A Messaging for Tasks:** Rejected because it lacks a persistent "Source of Truth" for late-joining teammates or recovery.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All task operations require a "Mission-Root Lineage" proof. Unauthorized agents cannot see or claim tasks outside their mission scope.
*   **Observability:** Integrated with the Teammate Task List Viewer for real-time mesh coordination tracking.

## 7. Evolutionary Changelog
*   **2026-07-08:** Initial Document Creation.
