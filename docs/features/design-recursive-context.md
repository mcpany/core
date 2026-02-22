# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-25

## 1. Context and Scope
As agents increasingly delegate tasks to subagents, there is no standardized way to pass the parent's context (goals, constraints, identity) down the chain. This leads to redundant prompts, loss of alignment, and "Context Bloat" as each subagent tries to re-establish the world state.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Define a standardized set of MCP headers for context inheritance.
    *   Implement a "Context Pruning" middleware that automatically filters the parent's context for the subagent's specific task.
    *   Enable subagents to report "State Delta" back to the parent.
*   **Non-Goals:**
    *   Creating a new inter-agent messaging protocol (we stay within MCP).
    *   Solving the "infinite recursion" problem automatically (requires policy).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Swarm Architect
*   **Primary Goal:** Have a "Research Agent" spawn a "Search Subagent" and ensure the subagent knows it must only look for academic sources.
*   **The Happy Path (Tasks):**
    1.  Parent agent initiates a tool call to a subagent.
    2.  MCP Any Middleware injects `X-MCP-Context-Parent-Goal` and `X-MCP-Context-Constraints` headers.
    3.  Subagent receives headers and incorporates them into its system prompt.
    4.  Subagent executes and returns results along with a `X-MCP-State-Update`.
    5.  Parent merges the state update into its own memory.

## 4. Design & Architecture
*   **System Flow:**
    `Parent Agent -> MCP Any (Middleware) -> Subagent -> MCP Any (Middleware) -> Parent Agent`
*   **APIs / Interfaces:**
    *   Standardized headers: `X-MCP-Context-*`
    *   New middleware: `ContextInheritanceMiddleware`
*   **Data Storage/State:**
    *   Temporary state stored in the Shared KV Store (Blackboard) if enabled.

## 5. Alternatives Considered
*   **Manual Prompt Injection:** Rejected as it is brittle and framework-specific. Headers provide a protocol-level solution.
*   **Vector Database for Context:** Good for large contexts, but headers are better for immediate constraints and goals.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Middleware must ensure that sensitive parent context (e.g., master API keys) is never leaked to subagents unless explicitly white-listed.
*   **Observability:** Context inheritance chains are visualized in the "Agent Black Box Player".

## 7. Evolutionary Changelog
*   **2026-02-25:** Initial Document Creation.
