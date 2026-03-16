# Design Doc: Distributed Supervisor Mesh (DSM) Orchestrator

**Status:** Draft
**Created:** 2026-05-08

## 1. Context and Scope
As enterprises transition from pilot projects to large-scale production swarms, the traditional "Central Supervisor" pattern—where a single orchestrator manages all subagents—has become a performance and reliability bottleneck. The Distributed Supervisor Mesh (DSM) evolves MCP Any from a central gateway into a decentralized coordination layer. It allows any agent within a swarm to act as a local supervisor for a sub-task while remaining cryptographically bound to the "Mission Root" established by the user.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable decentralized delegation of tasks between agents without a central bottleneck.
    * Maintain a cryptographically signed "Mission Root" that governs all sub-delegations.
    * Provide a unified audit trail for multi-hop delegations across the mesh.
    * Implement "Hierarchical Intent Leases" (HIL) to automatically revoke subagent capabilities upon task completion.
* **Non-Goals:**
    * Replacing the underlying LLM's reasoning capabilities.
    * Managing the transport layer for non-A2A compliant agents (handled by adapters).

## 3. Critical User Journey (CUJ)
* **User Persona:** Lead Systems Architect
* **Primary Goal:** Orchestrate a complex code migration across 50 repositories using a swarm of 100 agents without manual oversight of every sub-task.
* **The Happy Path (Tasks):**
    1. The user establishes a "Mission Root" for the code migration in MCP Any.
    2. The "Global Supervisor" agent delegates the "Frontend Migration" task to a "Sub-Supervisor" agent.
    3. MCP Any issues a Hierarchical Intent Lease to the Sub-Supervisor, derived from the Mission Root.
    4. The Sub-Supervisor further delegates component-level tasks to worker agents.
    5. MCP Any validates each worker agent's tool calls against the branched HIL.
    6. Upon completion of the component tasks, the worker leases expire; upon completion of the Frontend Migration, the Sub-Supervisor's lease is revoked.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root] -> [Primary Supervisor] --(Signed Intent)--> [Sub-Supervisor] --(Scoped Lease)--> [Worker Agents]`
* **APIs / Interfaces:**
    * `/v1/dsm/mission/create`: Establishes the Mission Root.
    * `/v1/dsm/delegate`: Validates a delegation request and issues a nested HIL.
    * `/v1/dsm/lease/status`: Monitors the health and expiration of active leases in the mesh.
* **Data Storage/State:**
    * The "Intent Graph" is stored in the Mesh-Aware Blackboard, mapping the hierarchy of supervisors and workers.

## 5. Alternatives Considered
* **Centralized Orchestration**: Rejected due to scalability issues and single-point-of-failure risks in production swarms.
* **Pure Peer-to-Peer (No Root)**: Rejected due to the lack of security governance and the risk of "Agentic Drift" where subagents diverge from user intent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All delegations must be cryptographically signed. Worker agents never receive "Raw" capabilities, only task-scoped leases derived from the parent.
* **Observability**: Visualized in the "DSM Delegation Graph" in the UI, showing the real-time hierarchy and lease status of the entire swarm.

## 7. Evolutionary Changelog
* **2026-05-08:** Initial Document Creation.
