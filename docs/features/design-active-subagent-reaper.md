# Design Doc: Active Subagent Reaper
**Status:** Draft
**Created:** 2026-04-03

## 1. Context and Scope
As AI agent swarms move toward deep, multi-threaded reasoning (e.g., OpenClaw v2.8), subagents are increasingly branched into speculative "What-if" paths. Current architectures (MCP, UAB) focus on *starting* agents but lack a robust protocol for *terminating* them when a branch is discarded. This leads to "Ghost Reasoning" where orphaned subagents continue to consume tokens and mutate shared state (Blackboard) after their primary mission is cancelled.

The Active Subagent Reaper is a mandatory lifecycle management layer in MCP Any that binds subagent existence to the cryptographic intent of the parent.

## 2. Goals & Non-Goals
* **Goals:**
    * Forcefully terminate subagent sessions (WebSocket, Process, or API-bound) when their parent intent branch is pruned.
    * Implement a "Heartbeat" protocol to detect silent subagent divergence or failure.
    * Automatically purge "Ghost State" (uncommitted Blackboard writes) from orphaned subagents.
* **Non-Goals:**
    * Managing OS-level processes not launched or proxied by MCP Any.
    * Providing a general-purpose job scheduler (this is intent-bound lifecycle only).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (e.g., OpenClaw Specialist Agent)
* **Primary Goal:** Prune a 5-agent reasoning branch without leaving "Zombie" agents consuming compute and injecting noise.
* **The Happy Path (Tasks):**
    1. Parent Agent creates a "Speculative Intent Branch" via UACO.
    2. MCP Any's Reaper registers the branch and assigns a `Branch-Lease-ID`.
    3. Three subagents are spawned, each providing a periodic `Heartbeat` token bound to the `Branch-Lease-ID`.
    4. Parent Agent determines the branch is a dead-end and sends a `PRUNE_INTENT` signal to MCP Any.
    5. The Reaper immediately invalidates the `Branch-Lease-ID`.
    6. MCP Any terminates the subagent connections and rolls back any uncommitted Blackboard writes associated with that lease.

## 4. Design & Architecture
* **System Flow:**
    * **Lease Registry**: An in-memory store tracking `Intent-ID -> {Subagent-Session-IDs, Expiry, Status}`.
    * **Heartbeat Collector**: A UDP/WebSocket listener that updates the `Last-Seen` timestamp for active leases.
    * **Reaper Daemon**: A background worker that sweeps the registry every 500ms for expired or explicitly pruned leases.
* **APIs / Interfaces:**
    * `POST /lifecycle/heartbeat`: `{ lease_id: string, signature: string }`
    * `DELETE /lifecycle/intent/{intent_id}`: Trigger manual pruning.
* **Data Storage/State:**
    * Uses a `Shadow-Table` in the Blackboard to hold "Dirty" writes from speculative branches until they are committed.

## 5. Alternatives Considered
* **Agent Self-Termination**: Rejected because compromised or "Hallucinating" agents cannot be trusted to kill themselves.
* **OS-Level SIGKILL**: Rejected as too blunt for API-based or containerized subagents; doesn't handle state cleanup.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Heartbeats must be signed by the subagent's session key. Pruning signals must be signed by the Parent Agent or an authorized Monitor.
* **Observability:** Logs all Reaper actions (Termination events, state rollbacks) for swarm forensics.

## 7. Evolutionary Changelog
* **2026-04-03:** Initial Document Creation.

### Update: 2026-04-04 - Lease-Bound Process Tree Isolation
**Context:** Today's market sync on "Cross-Framework State Leakage" highlights that WebSocket termination alone is insufficient if sub-processes (e.g., local python executors) remain active.
**Architecture Adjustment:**
* Extending the **Reaper Daemon** to track OS-level Process Groups (PGRPs) associated with a `Branch-Lease-ID`.
* Implementing mandatory "Namespace Pinning" for containerized subagents to ensure total resource isolation upon lease expiration.
**Security Impact:** Prevents "Dirty State" mutations from orphaned local executors that bypass A2A lifecycle hooks.
