# Design Doc: Context Delta Middleware
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rapid adoption of the Recursive Context Protocol (RCP), agent swarms are scaling to unprecedented depths. However, this has led to "Context Storms"—where redundant state inheritance consumes massive amounts of tokens and LLM context window space. Each subagent currently inherits the *entire* state of its parent, regardless of whether that state has changed.

The Context Delta Middleware aims to solve this by implementing a diff-based synchronization mechanism. Instead of passing full state, MCP Any will track state versions and only propagate the "deltas" (changes) required for the current execution scope.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Context Diffing" engine using JSON-Patch (RFC 6902) or similar.
    * Reduce token overhead for multi-agent handoffs by at least 50%.
    * Support "Context Checkpointing" for long-running agent sessions.
* **Non-Goals:**
    * Designing a new compression algorithm (focus on structural deltas).
    * Modifying the core LLM inference process.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Architect.
* **Primary Goal:** Efficiently synchronize state across a 10-agent deep hierarchy without hitting the 128k token limit.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a session and stores initial state.
    2. Parent agent modifies a subset of the state (e.g., updates a "research_objective" key).
    3. Subagent is spawned with a `X-MCP-Context-Version` header.
    4. MCP Any detects the version mismatch and computes the delta between the parent's current state and the subagent's known state.
    5. MCP Any injects only the delta into the tool execution context.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>MCP Any: Tool Call (Full State V1)
        MCP Any->>Delta Engine: Snapshot V1
        Agent A->>Agent B: Delegate (Token V1)
        Agent A->>MCP Any: Update State (V2 - Minor Change)
        Agent B->>MCP Any: Tool Call (Token V1)
        MCP Any->>Delta Engine: Compare V1 vs V2
        Delta Engine-->>MCP Any: JSON Patch (Delta)
        MCP Any->>Tool: Execute with Applied Delta
    ```
* **APIs / Interfaces:**
    * `GET /context/delta/:session_id?from=v1&to=v2`: Retrieve specific deltas.
    * Header: `X-MCP-Context-Delta: true`.
* **Data Storage/State:**
    * Versioned snapshots stored in the `Shared KV Store` (SQLite).

## 5. Alternatives Considered
* **Full Compression (Gzip/Brotli):** Rejected because LLMs cannot reason over compressed text; deltas must be readable JSON/Text.
* **Client-Side Diffing:** Rejected because it increases agent complexity; "dumb" agents should benefit from "smart" infrastructure.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Deltas must be cryptographically signed to prevent "Context Poisoning" (malicious delta injection).
* **Observability:** Tool Execution Timeline will show "Delta Sync" events with size reduction metrics.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation. Pivoting from recursive inheritance to delta-based synchronization to mitigate Context Storms.
