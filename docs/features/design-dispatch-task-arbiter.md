# Design Doc: Dispatch-Aware Task Arbiter (DATA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of "Dispatch & Channels" in Claude Code and similar parallel team coordination patterns in OpenClaw, agent swarms are increasingly operating in a high-concurrency, horizontal mode. However, this has introduced a critical pain point: **Cognitive Stall**. Parallel teammates often collide when attempting to claim tasks from a shared list, leading to extended wait cycles and synchronization deadlocks.

The Dispatch-Aware Task Arbiter (DATA) is designed to serve as the kernel-level coordination layer for these parallel swarms. By utilizing Conflict-Free Replicated Data Types (CRDTs), DATA enables teammates to claim, delegate, and update tasks asynchronously across multiple communication channels without global coordination locks.

## 2. Goals & Non-Goals
* **Goals:**
    * Resolve task-claiming conflicts across parallel dispatch lanes in sub-100ms.
    * Provide a framework-neutral coordination bus for Claude Code Channels and OpenClaw ACP swarms.
    * Eliminate "Cognitive Stall" by allowing agents to continue reasoning while state converges asynchronously.
    * Enforce mission-root priority during automated conflict resolution.
* **Non-Goals:**
    * Managing the internal reasoning state of individual agents.
    * Replacing the A2A Messaging Hub (DATA is a specialized coordination extension).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5 parallel Claude teammates across 3 dedicated Channels without "Task Shadowing" or wait-induced stalls.
* **The Happy Path (Tasks):**
    1. The Team Lead agent initializes a "Mission Backlog" via the DATA service.
    2. Teammates A and B simultaneously discover Task #42 on different Channels.
    3. Both agents issue a "Claim" intent to the DATA arbiter.
    4. DATA utilizes a Delta-CRDT LWW-Set to accept both claims locally.
    5. DATA resolves the collision using hardware-attested priority and monotonic timestamps.
    6. Teammate A is awarded Task #42; Teammate B receives an asynchronous "Pivot" signal and immediately moves to Task #43 without entering a wait state.

## 4. Design & Architecture
* **System Flow:**
    `[Teammate A] --(Claim)--> [DATA Arbiter (CRDT Sync)] <--(Claim)-- [Teammate B]`
                                    |
                                    v
                        [Conflict Resolution Engine]
                                    |
                                    v
                        [Broadcast Converged State]
* **APIs / Interfaces:**
    * `POST /v1/dispatch/claim`: Submit a task claim with channel-metadata.
    * `GET /v1/dispatch/backlog`: Retrieve the converged, non-blocking task list.
    * `WS /v1/dispatch/stream`: Real-time coordination stream for teammate-to-teammate state sync.
* **Data Storage/State:**
    * **In-memory Delta-CRDTs:** For high-frequency coordination.
    * **Blackboard Shards:** Periodic persistence of converged task states with mission-root anchors.

## 5. Alternatives Considered
* **Git-Based Locking (Claude Code Default):** Rejected as the primary mechanism due to the 2s+ filesystem latency, which is the root cause of "Cognitive Stall." DATA serves as the high-speed cache for this process.
* **Centralized SQL Locking:** Rejected due to performance overhead and the "Read Before Write" requirement which blocks parallel reasoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All DATA intents must be signed with a hardware-attested mission token. Unauthorized agents cannot "claim" or "evict" tasks.
* **Observability:** Real-time visualization via the "Teammate Task-List Arbiter Workspace" showing merge events and pivot signals.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
