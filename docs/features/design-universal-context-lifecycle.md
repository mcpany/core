# Design Doc: Universal Context Lifecycle Backend
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
With OpenClaw v2026.3.7 introducing a pluggable `ContextEngine`, there is a pressing need for a standardized backend that can manage the context lifecycle across different frameworks. Currently, context management (compression, summarization, retrieval) is fragmented. MCP Any can provide a "Universal Context Lifecycle" service that implements these hooks, ensuring that a subagent spawned in OpenClaw can seamlessly share or inherit context from a parent running in a different environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement the reference backend for OpenClaw's `ContextEngine` hooks (`bootstrap`, `ingest`, `assemble`, `compact`).
    * Provide a centralized "Context Store" that supports recursive inheritance.
    * Standardize context "compression" strategies to reduce token waste.
    * Enable cross-framework context sharing (e.g., OpenClaw subagent inheriting from a Claude Code parent).
* **Non-Goals:**
    * Replacing the LLM's internal KV cache.
    * Handling raw vector embeddings (this bridges to RAG, but isn't the primary goal).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Developer using heterogeneous frameworks.
* **Primary Goal:** Ensure a subagent has the "summarized intent" of the last 10 turns without passing the full transcript.
* **The Happy Path (Tasks):**
    1. Parent agent calls `Context.Ingest(turn_data)`.
    2. MCP Any runs the `compact` hook to generate a semantic summary.
    3. Subagent is spawned; MCP Any triggers `prepareSubagentSpawn`.
    4. Subagent calls `Context.Assemble()` and receives the optimized "Intent-Scoped" context.

## 4. Design & Architecture
* **System Flow:**
    `Agent Framework -> MCP Any Context Adapter -> [Lifecycle Hooks] -> Context Store (SQLite/Redis)`
* **APIs / Interfaces:**
    * `/context/bootstrap`
    * `/context/ingest`
    * `/context/assemble`
    * `/context/compact`
* **Data Storage/State:**
    * Leverages the `Shared KV Store` for persistence, but adds a structured "Lifecycle Layer" on top.

## 5. Alternatives Considered
* **In-Memory Framework Context:** Rejected as it doesn't allow cross-process or cross-framework sharing.
* **Simple Header Passing:** Rejected as it doesn't support active management (compression/summarization) during the lifecycle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Context is "Intent-Bound"; agents can only access context branches they are explicitly authorized for.
* **Observability:** UI provides a "Context Visualizer" showing how the context was compressed over time.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
