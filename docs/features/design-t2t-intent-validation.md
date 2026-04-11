# Design Doc: Teammate-to-Teammate (T2T) Intent Validation
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
The introduction of horizontal "Agent Teams" in Claude Code and OpenClaw has shifted the coordination bottleneck from hierarchical supervision to peer-to-peer message passing. Without validation, a compromised specialist agent can "Mailbox Inject" a teammate into unauthorized actions. T2T Intent Validation provides a high-speed, non-blocking middleware to ensure all peer instructions align with the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-speed validation layer for inter-agent coordinate messages (Mailboxes).
    * Ensure all T2T instructions carry a cryptographically signed mission-root token.
    * Use CRDTs for the shared task list to resolve "Mailbox Lock" bottlenecks.
    * Detect and block "Teammate Impersonation" attempts in horizontal meshes.
* **Non-Goals:**
    * Centralizing all teammate communication (validation is distributed/sharded).
    * Providing a general-purpose message bus for non-agent traffic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate between a "Frontend Agent" and a "Backend Agent" without allowing the Frontend Agent to bypass security gates via peer instructions.
* **The Happy Path (Tasks):**
    1. The "Frontend Agent" writes a task to the shared mailbox for the "Backend Agent."
    2. The T2T Middleware intercepts the message and extracts the attached mission-root token.
    3. The middleware validates that the requested action is within the scope authorized by the mission root.
    4. The "Backend Agent" receives the validated message and executes the task.
    5. Any state updates are synchronized via CRDT shards to avoid global locks.

## 4. Design & Architecture
* **System Flow:**
    `[Teammate A] -> (Encrypted Mailbox) -> [T2T Middleware] -> (Intent Validation) -> [Teammate B]`
* **APIs / Interfaces:**
    * `T2TMailbox`: `SendMessage(to AgentID, msg Message)`
    * `IntentArbiter`: `Validate(msg Message) (bool, error)`
* **Data Storage/State:**
    * Sharded, CRDT-based task lists stored in the Blackboard.

## 5. Alternatives Considered
* **Global Mailbox Lock:** Rejected due to high coordination latency (MTTC) in 3+ teammate swarms.
* **Hierarchical Proxying:** Rejected as it defeats the performance benefits of horizontal Agent Teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Peer trust is never implicit; every instruction must carry fresh hardware-attested intent.
* **Observability:** Tracked via the "Multi-Agent Swarm Topology Monitor" and T2T performance metrics.

## 7. Evolutionary Changelog
* **2026-04-11:** Initial Document Creation.
