# Design Doc: Neural Context Distiller
**Status:** Draft
**Created:** 2026-04-18

## 1. Context and Scope
As AI agent swarms grow in complexity and depth, the "Context Bloat" problem becomes a critical bottleneck. Agents quickly exhaust token limits, leading to "Lost in the Middle" errors and degraded reasoning. OpenClaw's NCC v1.0 has set a precedent for localized neural compression. MCP Any needs a framework-agnostic "Distiller" to provide these same benefits to all connected agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a pluggable middleware for semantic context distillation.
    * Support OpenClaw NCC v1.0 compression protocols.
    * Reduce context window usage by up to 90% while maintaining reasoning accuracy.
    * Provide a standardized interface for agents to request "State Distillation" before handoffs.
* **Non-Goals:**
    * Training a custom LLM for distillation (will use existing specialized models).
    * Modifying the agent's internal reasoning engine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Engineer
* **Primary Goal:** Successfully hand off a 128k token context from a "Research Agent" to a "Summarization Agent" without exceeding the target's 32k token window.
* **The Happy Path (Tasks):**
    1. The Research Agent completes its task and prepares to hand off state.
    2. The A2A Messaging Hub detects a context size mismatch.
    3. The Neural Context Distiller intercepts the handoff request.
    4. The Distiller applies the NCC v1.0 protocol to generate a semantically dense summary of the context.
    5. The distilled context (10k tokens) is signed and attached to the handoff token.
    6. The Summarization Agent receives the distilled state and maintains mission continuity.

## 4. Design & Architecture
* **System Flow:**
    `[Raw Context] -> [NCC Adapter] -> [Distillation Model] -> [Verifiable Summary] -> [Signed State Token]`
* **APIs / Interfaces:**
    * `DistillerService`: `Distill(context []Fragment, targetSize int) ([]Fragment, error)`
    * `NCCProvider`: Interface for external compression models.
* **Data Storage/State:**
    * Original and distilled fragments are stored in the Shard-Aware State Buffer.

## 5. Alternatives Considered
* **Heuristic Truncation**: Rejected as it causes "Context Amnesia" and breaks mission reasoning.
* **Large Context Models**: Not always available or cost-effective for high-frequency subagents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Distilled fragments must be cryptographically bound to the original source manifest.
* **Observability:** Distillation ratios and reasoning integrity scores are logged in the Telemetry Hub.

## 7. Evolutionary Changelog
* **2026-04-18:** Initial Document Creation.
