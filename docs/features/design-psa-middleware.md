# Design Doc: Proactive State Alignment (PSA) Middleware
**Status:** Draft
**Created:** 2026-03-29

## 1. Context and Scope
In complex multi-agent swarms, individual subagents maintain an internal reasoning state (monologue) that can diverge from the global mission state (Blackboard) during deep reasoning loops. This "Cognitive Drift" leads to inconsistent tool calls, redundant work, and eventual swarm divergence.

Proactive State Alignment (PSA) provides a background synchronization mechanism that continuously reconciles agent-local reasoning with the shared mission root. It ensures that every subagent operates on the most current "Truth" without requiring manual synchronization calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-frequency alignment heartbeat between subagents and the Shared KV Store.
    * Provide automated "Drift Detection" using semantic similarity scores.
    * Enable background state-patching to subagent context windows without interrupting the reasoning flow.
* **Non-Goals:**
    * Replacing the Blackboard as the primary source of truth.
    * Forcing agent termination on minor drift (focus is on correction, not just policing).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Prevent a "Refactoring Agent" from using outdated variable names that were changed by a "Naming Specialist" agent in a parallel branch.
* **The Happy Path (Tasks):**
    1. Agent A (Refactoring) starts a long-running reasoning task.
    2. Agent B (Naming) updates a key naming convention in the Blackboard.
    3. PSA Middleware detects the update and calculates the "Drift Score" for Agent A's current context.
    4. PSA injects a "State Alignment Fragment" into Agent A's next inference cycle.
    5. Agent A automatically adopts the new naming convention mid-task without a full context reset.

## 4. Design & Architecture
* **System Flow:**
    `[Blackboard Update] -> [PSA Monitor] -> [Semantic Compare] -> [Alignment Injection] -> [Subagent Inference]`
* **APIs / Interfaces:**
    * `POST /psa/align`: Manual trigger for state alignment.
    * `GET /psa/drift/{agent_id}`: Retrieve real-time drift metrics.
* **Data Storage/State:**
    * Uses the versioning history of the Shared KV Store to track delta changes.

## 5. Alternatives Considered
* **Reactive Sync (On-Demand):** Rejected as it only detects drift *after* an inconsistent tool call is made.
* **Global Context Broadcast:** Rejected due to token window exhaustion in high-density swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Alignment fragments must be signed by the PSA Hub to prevent "State Injection" attacks.
* **Observability:** PSA events (alignments, corrections) are logged to the Swarm Topology Monitor.

## 7. Evolutionary Changelog
* **2026-03-29:** Initial Document Creation.
