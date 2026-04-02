# Design Doc: Speculative State Peering (SSP) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward horizontal, multi-teammate coordination (e.g., Claude Code Agent Teams), the primary performance and stability bottleneck has become the "Cognitive Stall." This occurs when parallel teammates encounter state conflicts on the Shared Blackboard or mailbox only *after* completing significant reasoning work, leading to expensive rollbacks and multi-second wait cycles.

MCP Any needs to solve this by providing a mechanism for agents to peer into each other's "Draft Reasoning" or "Speculative State" before a task claim or state mutation is finalized. The SSP Broker provides the secure, hardware-attested infrastructure for this pre-commit transparency.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable sub-100ms visibility into "Draft" reasoning fragments of parallel teammates.
    * Mandate hardware-attestation for all peered speculative buffers.
    * Provide "Conflict-Ahead" signals to agents to prevent divergent reasoning paths.
    * Ensure speculative peering does not lead to premature "Intent Leakage" to unauthorized subagents.
* **Non-Goals:**
    * This system WILL NOT automatically resolve conflicts; it only provides the visibility for agents to self-correct.
    * This system WILL NOT allow speculative state to be committed to the global Blackboard without passing standard quorums.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (running 10+ parallel teammates)
* **Primary Goal:** Identify a conflict between Teammate A's proposed file edit and Teammate B's proposed refactor before either agent spends 2000+ tokens on the task.
* **The Happy Path (Tasks):**
    1. Teammate A begins reasoning on a task and publishes a "Speculative Fragment" to the SSP Broker.
    2. Teammate B, before starting its own task, queries the SSP Broker for peer fragments in the same mission-root namespace.
    3. The SSP Broker verifies Teammate B's hardware-attested mission token.
    4. The SSP Broker provides Teammate B with a minimized, semantically-redacted view of Teammate A's draft.
    5. Teammate B identifies a potential conflict and triggers a "Speculative Reflection" loop to adjust its path or signal a negotiation.

## 4. Design & Architecture
* **System Flow:**
    * Teammate -> Publish(Fragment, HardwareToken) -> SSP Broker (In-Memory Buffer)
    * Peer Teammate -> Query(Namespace, HardwareToken) -> SSP Broker -> Verify(Token) -> Filter/Redact -> Peer Teammate
* **APIs / Interfaces:**
    * `POST /v1/ssp/publish`: Payload includes `fragment`, `mission_root_id`, and `attestation_receipt`.
    * `GET /v1/ssp/peer`: Returns a list of active speculative fragments for a given `mission_root_id`.
* **Data Storage/State:**
    * High-speed, in-memory LRU cache for speculative fragments.
    * Fragments are TTL-bound and automatically purged upon mission-step completion or teammate termination.

## 5. Alternatives Considered
* **Asynchronous Mailbox Sharding (AMS) only**: Rejected because AMS only handles finalized coordination messages. It does nothing to prevent "Reasoning Waste" on conflicting paths that haven't been published yet.
* **Global Locking**: Rejected due to prohibitive latency (2s+ stalls) and inability to scale with high-density swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * All fragments are encrypted at rest in memory.
    * Access requires hardware-attested mission-lineage tokens (RMRA compliant).
    * Semantic Redaction (ARR Hub v2) is applied to fragments before peering to prevent "Monologue Smearing."
* **Observability:**
    * "Speculative Conflict" alerts are surfaced to the UI (Parallel Intent Visualizer).
    * Metrics track "Reasoning Saved" by calculating token delta between conflict detection and original estimated task depth.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
