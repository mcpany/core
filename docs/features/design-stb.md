# Design Doc: Speculative Task Broker (STB)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As agent meshes grow in complexity, the latency of hardware-attested identity handshakes (often >100ms) has become a primary bottleneck for real-time coordination. OpenClaw v3.5.0 has introduced "Speculative Coordination," where agents can begin reasoning against a task before their identity is fully verified.

The Speculative Task Broker (STB) in MCP Any provides the necessary infrastructure for this speculative execution. It hosts "Speculative State Buffers"—isolated, transactional memory regions where an agent's mutations (to the Blackboard or Context Shards) are stored temporarily. These mutations are only committed to the mission-root state once a valid attestation signal is received; otherwise, they are atomically rolled back.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide transactional isolation for speculative agent mutations.
    * Implement atomic `Commit` and `Rollback` triggers bound to FSI/HADH attestation results.
    * Neutralize "Speculative State Poisoning" by ensuring zero visibility of speculative fragments to other non-speculative teammates.
    * Minimize coordination latency by allowing reasoning and attestation to proceed in parallel.
* **Non-Goals:**
    * Executing the agent's reasoning (STB is a state mediator).
    * Providing long-term persistence for un-committed speculative state.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-frequency Specialist Agent
* **Primary Goal:** Begin processing a time-critical data refactoring task while the 150ms identity handshake completes.
* **The Happy Path (Tasks):**
    1. Specialist Agent receives a task proposal and issues a `Speculative Claim` to the STB.
    2. STB opens a `Speculative Shard` in the BSH State Buffer.
    3. Agent performs multiple tool calls, writing results to the Speculative Shard.
    4. Simultaneously, the FSI Provider completes the hardware-attested handshake.
    5. FSI issues a `Success` signal to the STB.
    6. STB atomically merges the Speculative Shard into the mission-root Blackboard.
    7. Specialist Agent is notified that its state is now "Durable."

## 4. Design & Architecture
* **System Flow:**
    `Task Proposal` -> `Speculative Claim (STB)` -> `Reasoning (Parallel)` | `Attestation (Parallel)` -> `Commit/Rollback`
* **APIs / Interfaces:**
    * `STB.OpenBuffer(session_id string) (buffer_id string, err error)`
    * `STB.Commit(buffer_id string) error`
    * `STB.Rollback(buffer_id string) error`
* **Data Storage/State:**
    * Utilizes `memfd_create` for high-speed, isolated memory-mapped buffers for speculative state fragments.

## 5. Alternatives Considered
* **Synchronous Handshaking**: Rejected due to high "Cognitive Stall" in deep swarms.
* **Optimistic Mainline Commit**: Rejected as it allows un-attested subagents to poison the mission-root state before verification.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative state is strictly isolated from the mainline until hardware attestation is confirmed.
* **Observability:** Speculative cycles and rollback rates are tracked in the "Speculative Coordination Monitor."

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
