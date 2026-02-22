# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agent swarms become more complex, subagents often need access to a subset of the parent agent's context (e.g., specific variables, recent history, or task goals). Current MCP implementations often pass the entire context, leading to "context bloat" and token wastage. MCP Any needs a standardized way to handle partial, recursive context inheritance.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize headers for context inheritance between agents.
    * Allow agents to define "scoped" context fragments.
    * Reduce token usage by pruning irrelevant context for subagents.
* **Non-Goals:**
    * Implementing a full state-machine for agents.
    * Handling real-time synchronisation of context (this is covered by the Shared KV Store).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator
* **Primary Goal:** Pass only the "Database Schema" context to a "SQL Expert" subagent without passing the entire user conversation history.
* **The Happy Path (Tasks):**
    1. Parent agent identifies the need for a SQL subagent.
    2. Parent agent wraps the required schema in a `<mcp:context id="schema">` tag.
    3. MCP Any core intercepts the subagent call and extracts only the tagged context.
    4. Subagent receives a lean prompt containing only the necessary schema.

## 4. Design & Architecture
* **System Flow:** Parent -> Tool Call (with context tags) -> MCP Any (Extraction) -> Subagent.
* **APIs / Interfaces:** New `mcp_context` metadata field in `CallToolRequest`.
* **Data Storage/State:** Transient, scoped to the lifecycle of the tool call.

## 5. Alternatives Considered
* **Passing everything:** Rejected due to cost and performance (hallucinations increase with context size).
* **Manual filtering in agent prompts:** Rejected as it places too much burden on the model to "correctly" filter its own context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Ensure that sensitive context (e.g., auth tokens) is explicitly marked and not inherited by default.
* **Observability:** Log the "context pruning" ratio to demonstrate token savings.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
