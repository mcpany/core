<!-- markdownlint-disable -->

# Design Doc: LFMC (Lock-Free Mesh Coordination)

**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope

In large, horizontal agent swarms, "Mailbox Lock" bottlenecks prevent parallel teammates from coordinating at machine speed. LFMC introduces a lock-free, CRDT-based coordination bus that allows teammates to update their state fragments without global synchronization.

## 2. Goals & Non-Goals

* **Goals:**

    * Eliminate coordination-layer bottlenecks in large swarms.

    * Provide "Eventually Consistent" state for teammate task lists.

    * Support "Sovereign Sharding" of context fragments.

* **Non-Goals:**

    * Strict consistency for mission-critical intents (handled by HAIL).

## 3. Critical User Journey (CUJ)

* **User Persona:** Swarm Orchestrator

* **Primary Goal:** Coordinate 10 specialists on a complex refactor without global lock contention.

* **The Happy Path (Tasks):**

    1. Orchestrator initializes an LFMC mesh for the mission.
    2. Specialists claim tasks in their "Sovereign Shards."
    3. Fragment updates are propagated asynchronously via the mesh bus.
    4. LFMC reconciles concurrent updates using "Semantic Merging."
    5. Team reaches eventually consistent state at sub-10ms latency.

## 4. Design & Architecture

* **System Flow:**

    `[Specialist A (Shard)] <-> [LFMC Bus (CRDT)] <-> [Specialist B (Shard)]`

* **APIs / Interfaces:**

    * `/v1/mesh/shard`: Endpoint for fragment submission.

    * `/v1/mesh/sync`: High-speed synchronization channel (Named Pipe).

* **Data Storage/State:** Mesh state is stored in a distributed, lock-free Blackboard using CRDT primitives.

## 5. Alternatives Considered

* **Global Redis Lock**: Rejected due to latency and the "Stale Shard" problem in high-frequency swarms.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Every shard update must be HAIL-attested to prevent "Context Smearing" by malicious teammates.

* **Observability:** Convergence latency and shard-collision rates are monitored.

## 7. Evolutionary Changelog

* **2026-06-19:** Initial Document Creation.
