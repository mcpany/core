# Design Doc: Universal Multimodal Memory Bus (UMMB)
**Status:** Draft
**Created:** 2026-07-01

## 1. Context and Scope
With the rise of horizontal teammate coordination (e.g., Claude Code Agent Teams) and multimodal agentic workflows (NVIDIA Agent Toolkit), the traditional "Parent-Mediated" state relay has become a critical bottleneck. Agents in a swarm currently lack a direct, high-performance way to synchronize complex, multimodal state (images, audio traces, SVG fragments) without bloating the parent's context window or incurring significant relay latency.

The **Universal Multimodal Memory Bus (UMMB)** is a hardware-attested, intent-pinned coordination layer for MCP Any. It provides a "Shared Memory" architecture for swarms, allowing parallel teammates to synchronize state shards directly while ensuring the absolute sovereignty of the mission root through real-time multimodal trace sanitization.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a low-latency, lock-free memory bus for inter-agent state synchronization.
    * Support "Multimodal Sharding" (images, audio, and structured data) within the Shared KV Store.
    * Implement real-time "Multimodal Trace Sanitization" to prevent context-smuggling via non-textual payloads.
    * Mandate hardware-attested "Intent-Pinning" for every memory shard to ensure mission-root alignment.
* **Non-Goals:**
    * Replacing long-term archival storage (UMMB is for active session coordination).
    * Providing a general-purpose file sharing system (it is limited to agent-reasoning state).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multimodal Swarm Orchestrator
* **Primary Goal:** Synchronize a UI screenshot and its corresponding "Self-Correction" reasoning between a "Designer Agent" and a "QA Agent" without relaying through the "Project Manager."
* **The Happy Path (Tasks):**
    1. The Designer Agent generates a UI fragment and pushes the `image/png` shard to the UMMB.
    2. The UMMB performs a "Multimodal Integrity Scan" (MITS) to ensure no hidden instructions are in the metadata.
    3. The Designer Agent tags the shard with the hardware-attested `Mission-Root-Token`.
    4. The QA Agent, also bound to the same mission, "mounts" the shard instantly from the UMMB.
    5. The QA Agent performs its review and pushes a "Correction Shard" back to the bus.
    6. Both agents remain synchronized in sub-100ms, while the Project Manager's context remains focused on high-level orchestration.

## 4. Design & Architecture
* **System Flow:**
    `Agent A` -> `UMMB Producer` -> `MITS Sanitizer` -> `Intent-Pinned Shard (Blackboard)` <- `UMMB Consumer` <- `Agent B`
* **APIs / Interfaces:**
    * `UMMB.publishShard(type, payload, mission_token)`
    * `UMMB.subscribeShard(shard_id, mission_token)`
    * `UMMB.verifyLineage(shard_id)`
* **Data Storage/State:**
    * Utilizes the **Mesh-Aware Blackboard** as the underlying storage engine, extended with binary blob support for multimodal fragments.

## 5. Alternatives Considered
* **Parent-Relay (Status Quo):** Rejected due to the "Relay Bottleneck" and context-window bloat in the parent agent.
* **S3/Cloud Storage:** Rejected due to high latency and the lack of integrated "Intent-Pinning" at the storage layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All shards are cryptographically bound to the mission-root. Multimodal payloads undergo mandatory MITS (Multimodal Inference-Time Sanitization).
* **Observability:** Shard lineage and "Attention-Density" metrics are exposed to the Visual Attention Dashboard.

## 7. Evolutionary Changelog
* **2026-07-01:** Initial Document Creation.
