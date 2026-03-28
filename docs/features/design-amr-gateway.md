# Design Doc: Atomic Mission-Resumption (AMR) Gateway
**Status:** Draft
**Created:** 2026-06-24

## 1. Context and Scope
As agent missions become multi-day and span hundreds of tool calls, the cost of "Cold Booting" a session (re-injecting 100k+ tokens of reasoning history) becomes prohibitive in terms of both latency and token economics. Currently, if an orchestration node crashes or a teammate rotates, the agent must often re-read the entire context to reach the same cognitive state.

The **Atomic Mission-Resumption (AMR) Gateway** provides a hardware-locked mechanism to snapshot the "Reasoning Frontier." It allows agents to resume execution from a verified, cryptographically signed point-in-time, ensuring that "Cognitive Stall" is eliminated during process handoffs or recovery events.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate sub-100ms resumption of agent reasoning states across cold-boots.
    * Mandate hardware-bound (TPM/Secure Enclave) signatures for all Mission Snapshots.
    * Ensure bit-perfect state consistency for Binary State Handoffs (BSH).
* **Non-Goals:**
    * Providing a long-term archival service for all historical reasoning traces.
    * Implementing natural-language diffing between reasoning snapshots.
    * Managing model-specific KV-cache persistence (handled by providers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Infrastructure Architect
* **Primary Goal:** Recover a distributed swarm mission after a cluster-level restart without losing mission-root lineage.
* **The Happy Path (Tasks):**
    1. The parent agent establishes a "Mission Root" through the AMR Gateway.
    2. During execution, the Gateway periodically generates "Atomic Snapshots" of the Blackboard and reasoning frontier.
    3. The orchestrator node experiences a failure and restarts.
    4. Upon reboot, the agent requests mission resumption using a hardware-attested "Resumption Token."
    5. The AMR Gateway verifies the token signature against the TPM and injects the "Frontier Shard" back into the agent's context.
    6. The agent resumes reasoning immediately from the last verified turn.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [Tool Call] -> AMR Gateway (Signs & Stores Snapshot) -> MCP Server`
    `Orchestrator (Reboot) -> AMR Gateway (Verifies Token) -> Agent (Injected State)`
* **APIs / Interfaces:**
    * `mcp.amr.v1.SnapshotMission(mission_id, state_blob) -> resumption_token`
    * `mcp.amr.v1.ResumeMission(resumption_token) -> mission_state`
* **Data Storage/State:**
    * Snapshots are stored as Protobuf-encoded BSH fragments.
    * Keys are managed via the **Mesh-Resident Key Exchange (MRKE)** provider.

## 5. Alternatives Considered
* **JSON-based Checkpointing:** Rejected due to the "Token Storm" overhead and lack of fragment-level integrity.
* **Persistent External DB (Postgres/Redis):** Rejected as the primary transport due to the latency tax of serialization. BSH + Shared Memory is preferred for "Atomic" speed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All snapshots must be signed by the originating agent's hardware identity. Unauthorized resumption requests trigger immediate "Sovereignty Corruption" signals.
* **Observability:** We will track `SnapshotHitRate` and `ResumptionLatencyMs` to monitor gateway efficiency.

## 7. Evolutionary Changelog
* **2026-06-24:** Initial Document Creation.
