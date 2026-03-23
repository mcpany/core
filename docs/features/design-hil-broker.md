# Design Doc: Hierarchical Intent Lease (HIL) Broker
**Status:** Draft
**Created:** 2026-05-03

## 1. Context and Scope
As agent swarms grow in complexity, the "Implicit Trust" of subagents becomes a major liability. The current "Session-Bound" privilege model is too coarse, allowing subagents to retain capabilities long after their specific sub-task is complete. The Hierarchical Intent Lease (HIL) Broker implements the UACO v3.2 standard for task-bound, hierarchical capability management. This ensures that every subagent's privileges are cryptographically tied to a specific node in the "Intent Tree" and are automatically revoked upon sub-mission completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement UACO v3.2 Hierarchical Intent Leases.
    * Provide cryptographic binding between subagent capabilities and specific sub-tasks.
    * Enable automated "Cascading Revocation" where pruning a parent intent automatically clears all child leases.
    * Integrate with the HAFP (Hardware-Bound Fast-Path) for low-latency lease validation.
* **Non-Goals:**
    * Defining the business logic for task branching (handled by the Agent Framework).
    * Providing long-term persistent storage for leases (leases are ephemeral by design).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Grant a "Read-Only FS" lease to a "Researcher Subagent" that only exists for the duration of the "Literature Review" sub-task.
* **The Happy Path (Tasks):**
    1. Parent Agent defines a "Literature Review" sub-task.
    2. HIL Broker generates a Task-Bound Lease tied to this sub-task ID.
    3. Researcher Subagent receives the lease and can access the FS tool.
    4. Upon sub-task completion signal (UACO v3.2), the HIL Broker invalidates the lease.
    5. Subsequent FS tool calls by the Researcher are rejected.

## 4. Design & Architecture
* **System Flow:**
    `Task Proposal` -> `Lease Generation` -> `Capability Mapping` -> `Mission Execution` -> `Completion Signal` -> `Automated Revocation`
* **APIs / Interfaces:**
    * `HILBroker`: Core service for issuing and verifying hierarchical leases.
    * `IntentTreeWatcher`: Monitors the Blackboard for sub-task lifecycle changes.
    * `LeaseEnforcer`: Middleware that validates tool calls against active HIL tokens.
* **Data Storage/State:**
    * Leases are stored in an in-memory "Lease Registry," indexed by Task ID and cryptographically linked to the Root Mission Intent.

## 5. Alternatives Considered
* **Time-Bound Leases (TTL)**: Rejected because tasks often have variable durations, leading to either premature expiration or "zombie" privileges.
* **Manual Revocation**: Rejected as it is too slow and error-prone for high-velocity agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are signed by the HIL Broker's hardware key and carry a "Context Proof" of their parentage.
* **Observability:** Active leases and revocation events are visualized in the "Hierarchical Trust Monitor."

## 7. Evolutionary Changelog
* **2026-05-03:** Initial Document Creation.
