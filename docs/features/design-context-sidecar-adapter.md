# Design Doc: Context Sidecar Adapter
**Status:** Draft
**Created:** 2026-04-16

## 1. Context and Scope
As agent frameworks like OpenClaw and AutoGen evolve, they are developing increasingly sophisticated and proprietary context management strategies. This leads to "Context Silos," where critical state is trapped within a single framework. The `Context Sidecar Adapter` is designed to break these silos by providing a standardized "Context Bus" that allows MCP Any to host and bridge these framework-specific strategies as pluggable sidecars.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized interface for external "Context Sidecars" to plug into MCP Any.
    * Enable real-time state synchronization between disparate agent frameworks.
    * Support OpenClaw's "Contextual Anchor" standard for mission-pinning.
    * Implement a "Verifiable State Handoff" protocol using the Delegation Attestation Layer.
* **Non-Goals:**
    * Replacing the internal memory of agent frameworks.
    * Forcing a single, monolithic context schema (sidecars can use their own internal formats).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Share mission-critical "Anchored Context" between an OpenClaw research agent and an AutoGen coding subagent.
* **The Happy Path (Tasks):**
    1. The OpenClaw agent registers its "Contextual Anchor" with the Context Sidecar Adapter.
    2. The Adapter stores the anchor in the Shared Blackboard, cryptographically bound to the Mission Intent.
    3. When the AutoGen subagent is triggered, the Adapter automatically injects the "Anchored Context" into the subagent's prompt prefix.
    4. The Adapter monitors both agents' state via their respective sidecars, ensuring no "Intent Ghosting" occurs during parallel branches.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Framework] <-> [Framework Sidecar (WASM/gRPC)] <-> [Context Sidecar Adapter] <-> [Shared Blackboard]`
* **APIs / Interfaces:**
    * `sidecar/register`: Register a new framework-specific sidecar.
    * `sidecar/anchor`: Pin a mission-critical context fragment.
    * `sidecar/sync`: Synchronize state across registered sidecars.
* **Data Storage/State:**
    * State is persisted in the existing SQLite-based Shared Blackboard.
    * Cryptographic intent-binding ensures state is only accessible within the authorized mission scope.

## 5. Alternatives Considered
* **Centralized Vector Store:** Rejected for real-time inter-agent handoffs due to latency and lack of intent-binding.
* **Manual Context Mirroring:** Rejected as it is error-prone and doesn't scale for complex swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All state handoffs must be verified by the Delegation Attestation Layer. Context is isolated by default and only shared via explicit "Anchors."
* **Observability:** Integrated with the "Context Sidecar Sync Viewer" UI to visualize state flow and anchor status in real-time.

## 7. Evolutionary Changelog
* **2026-04-16:** Initial Document Creation.
