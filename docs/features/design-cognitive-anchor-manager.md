# Design Doc: Cognitive Anchor Manager
**Status:** Draft
**Created:** 2026-04-27

## 1. Context and Scope
As AI agent swarms grow in complexity and reasoning depth, they often suffer from "Semantic Drift"—a phenomenon where subagents lose sight of the primary mission intent as they descend into recursive task execution. OpenClaw v2026.3.8 introduced "Cognitive Anchoring" to mitigate this by pinning high-level mission goals within the context engine.

MCP Any, as the universal infrastructure layer, must provide a standardized way to host and manage these anchors across disparate agent frameworks. The Cognitive Anchor Manager (CAM) will act as an extension of the ContextEngine Adapter, ensuring that mission-root intents remain immutable and persistent throughout the multi-agent reasoning lifecycle.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure registry for "Cognitive Anchors" (immutable intent fragments).
    * Support "Root-Anchor Persistence" across deep agent handoffs.
    * Implement "Adaptive Anchor Pruning" to prevent context bloat (OpenClaw v2026.3.9 compliance).
    * Ensure hardware-bound cryptographic protection for mission-root anchors.
* **Non-Goals:**
    * Perform the LLM reasoning itself (the LLM remains the reasoner).
    * Replace the primary agent context (CAM is a sidecar/specialized manager).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Maintain mission consistency across a 5-level deep agent hierarchy without "Intent Ghosting."
* **The Happy Path (Tasks):**
    1. Parent agent initializes a "Mission Root Anchor" via MCP Any CAM API.
    2. Parent agent delegates a task to a subagent, including the Root Anchor ID.
    3. MCP Any intercepts the subagent's tool calls and automatically injects the Root Anchor into the retrieval context.
    4. As sub-subagents generate secondary anchors, CAM performs "Adaptive Pruning" to keep the context window optimized while keeping the Root Anchor pinned.
    5. User verifies the reasoning path in the A2UI, seeing the mission intent consistently applied.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|Tool Call / State Sync| Server[MCP Any Server]
        Server -->|Middleware| CAM[Cognitive Anchor Manager]
        CAM -->|Pruning Logic| Engine[Smart Pruning Engine]
        CAM -->|Storage| Vault[Immutable Anchor Vault]
        CAM -->|Context Injection| Context[ContextEngine Adapter]
        Context -->|Retrieval| Agent
    ```
* **APIs / Interfaces:**
    * `POST /v1/anchors`: Create a new cognitive anchor.
    * `GET /v1/anchors/{id}`: Retrieve anchor metadata and content.
    * `PUT /v1/anchors/{id}/bind`: Bind an anchor to a specific session or intent-scope.
* **Data Storage/State:**
    * Anchors are stored in an internal SQLite-backed "Immutable Vault" with SHA-256 integrity checks.

## 5. Alternatives Considered
* **Client-Side Anchoring:** Rejected because it relies on the honesty of the agent framework and increases "Token Storm" overhead if every agent must re-send anchors.
* **Monolithic Context Engines:** Rejected because it lacks the flexibility to bridge between OpenClaw, AutoGen, and Gemini's native context formats.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mission-root anchors are cryptographically signed. Subagents can add "Branch Anchors" but cannot mutate the "Root Anchor" without parent-level re-attestation.
* **Observability:** CAM will export "Alignment Scores" showing how closely active reasoning fragments align with the pinned anchors.

## 7. Evolutionary Changelog
* **2026-04-27:** Initial Document Creation.
* **2026-04-27 Update:** Integrated "Smart Pruning Middleware" and "LFV Receipt" requirements based on today's market sync.
