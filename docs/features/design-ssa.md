# Design Doc: Shared State Arbiter (SSA)
**Status:** Draft
**Created:** 2026-03-21

## 1. Context and Scope
As AI agent swarms transition from vertical hierarchies to horizontal "Agent Teams" (e.g., Claude Code, CrewAI), specialized subagents increasingly compete for shared resources within the MCP Any Blackboard (Shared KV Store). Without central arbitration, parallel reasoning paths frequently enter "Reasoning Loops" or deadlocks, where Agent A waits for a state change from Agent B, who is simultaneously blocked by Agent A.

The Shared State Arbiter (SSA) is a core orchestration service for the Blackboard that provides real-time wait-graph analysis and authoritative deadlock resolution. It moves MCP Any from a passive storage layer to an active participant in swarm stability.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized "Wait-Graph" to track resource dependencies between session-bound agents.
    * Detect circular dependencies (deadlocks) in multi-agent Blackboard transactions.
    * Provide a "Checkpoint-and-Yield" API allowing high-priority agents to preempt resource locks.
    * Standardize "Reasoning-Aware" timeouts to prevent infinite stalling in horizontal swarms.
* **Non-Goals:**
    * Replacing the underlying storage engine (SQLite/Redis) of the Blackboard.
    * Managing inter-agent message passing (handled by A2A Messaging Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Resolve a circular dependency between a "Coder" agent and a "Reviewer" agent both trying to update the same project-local config state.
* **The Happy Path (Tasks):**
    1. Coder Agent requests an exclusive lock on `kv://project/config/status`.
    2. Reviewer Agent requests the same lock and is put into a "WAITING" state by SSA.
    3. SSA adds a directed edge `Reviewer -> Coder` to the Wait-Graph.
    4. Coder Agent then requests a read-lock on `kv://project/review/results` currently held by Reviewer.
    5. SSA detects a cycle: `Coder -> Reviewer -> Coder`.
    6. SSA triggers the Deadlock Resolver, which evaluates "Mission Lineage" and instructs the Reviewer to "Yield-and-Rollback."
    7. Reviewer Agent releases its lock and checkpoints its current reasoning state.
    8. Coder Agent completes its transaction; Reviewer automatically resumes.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Agent A] -> [Blackboard API] -> [SSA Lock Manager] -> [Wait-Graph Store]
                                            |
                                   [Deadlock Detector] -> [Mission Policy Engine]
                                            |
    [Agent B] <- [Yield Signal] <-----------+
    ```
* **APIs / Interfaces:**
    * `POST /v1/ssa/lock`: Request a scoped lock with priority and lineage headers.
    * `GET /v1/ssa/graph`: Debug endpoint to visualize the current dependency mesh.
    * `POST /v1/ssa/yield`: Explicitly signal a willingness to yield state for higher-priority intents.
* **Data Storage/State:**
    * SSA maintains an in-memory directed graph of `(AgentID, ResourceID)` pairs.
    * Persistent state is minimal, as SSA focuses on the *active reasoning session*.

## 5. Alternatives Considered
* **First-Come-First-Served (FCFS) Locking:** Rejected because it leads to "Reasoning Starvation" where a slow-reasoning subagent blocks the primary mission root indefinitely.
* **Distributed Locking (Consul/Etcd):** Rejected as too heavyweight for local-first agentic infrastructure; MCP Any needs sub-millisecond arbitration latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SSA verifies that yield signals and lock requests are bound to hardware-attested session tokens (from the Handshake Provider).
* **Observability:** SSA logs all wait-graph "Breaks" (forced yields), providing critical data for RL-driven swarms to optimize their coordination logic.

## 7. Evolutionary Changelog
* **2026-03-21:** Initial Document Creation.
