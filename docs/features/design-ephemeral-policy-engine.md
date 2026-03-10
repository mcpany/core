# Design Doc: Ephemeral Policy Engine
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the rise of "Ephemeral Tooling" (LLM-generated tools), MCP Any needs a policy engine that can evaluate dynamic tool schemas and runtime-defined instructions without introducing latency. Current static Rego-based policies are designed for well-defined, persistent APIs. Ephemeral tools require a more agile and performant evaluation runtime that can handle rapidly changing tool definitions.

## 2. Goals & Non-Goals
* **Goals:**
    * Create a JIT-compiled policy runtime for dynamic tool evaluation.
    * Support "Zero-Knowledge" schemas where tool details are verified against intent-scoped patterns.
    * Optimize for extremely low-latency (< 5ms) policy checks on ephemeral tool calls.
    * Integrate with existing Rego/CEL policies for hybrid (static + dynamic) enforcement.
* **Non-Goals:**
    * Replace all existing Rego policies for static services.
    * Automate the generation of policies for all possible LLM tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using Claude Code with ephemeral tool generation.
* **Primary Goal:** Safely execute a runtime-defined tool that processes local data without exposing the entire filesystem.
* **The Happy Path (Tasks):**
    1. Claude Code generates an ephemeral tool (e.g., `process_local_csv`).
    2. The tool definition is "registered" with MCP Any as an ephemeral service.
    3. When the agent calls `process_local_csv`, the `Ephemeral Policy Engine` JIT-compiles a validator based on the current "Intent Scope" (e.g., "process only files in /tmp/data").
    4. The engine validates the call parameters against this intent.
    5. The tool is executed in a detached sandbox if the policy passes.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|Call Ephemeral Tool| Core[MCP Any Core]
        Core -->|Tool Schema + Call Params| EPE[Ephemeral Policy Engine]
        EPE -->|JIT Compile| Policy[Active Policy Guard]
        Policy -->|Allow/Deny| Core
        Core -->|Execute| Sandbox[Detached Sandbox]
        Sandbox -->|Result| Core
        Core -->|Result| Agent
    ```
* **APIs / Interfaces:**
    * `EphemeralPolicyRuntime` interface with `Evaluate(toolSchema, callParams)` method.
    * JIT-compilation logic for dynamic schema validation.
* **Data Storage/State:**
    * In-memory cache for compiled policy "hot-spots."

## 5. Alternatives Considered
* **Interpreted Rego on every call:** Rejected due to performance overhead for large swarms and dynamic schemas.
* **Stateless Policy Workers:** Rejected as they would increase network latency for every tool call.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mitigates risks from LLM-generated tools that might attempt to perform unauthorized actions.
* **Observability:** Detailed logs of JIT-compilation times and policy hits/misses.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
