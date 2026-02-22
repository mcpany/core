# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents evolve from single-task bots into complex, multi-agent swarms, the ability to share context across generations (Parent -> Child -> Subagent) becomes critical. Currently, MCP (Model Context Protocol) focuses on the interaction between a single client and a server. When an agent (acting as a client) spawns another agent, the "intent," "security constraints," and "session history" of the parent are often lost or must be manually re-injected.

MCP Any needs to solve this by implementing a middleware layer that standardizes context inheritance, ensuring that every tool call in a swarm carries the weight of the overarching mission.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Standardize a `X-MCP-Context-Parent` header for all JSON-RPC calls.
    *   Implement a middleware that automatically injects/extracts this context.
    *   Provide a "Context Store" where subagents can query parent state.
*   **Non-Goals:**
    *   Automatically merging chat histories (too model-specific).
    *   Solving multi-modal context (images/video) in the first iteration.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator (e.g., OpenClaw user).
*   **Primary Goal:** Subagent executes a tool call with the same security scope and high-level goal as the parent agent without manual propagation.
*   **The Happy Path (Tasks):**
    1.  Parent Agent initiates a session with a `mission_id` and `security_policy`.
    2.  Parent Agent spawns a "Researcher" Subagent via a tool call.
    3.  MCP Any intercepts the call, generates a child context, and attaches the `mission_id`.
    4.  Researcher Subagent calls `search_web`.
    5.  `search_web` adapter verifies the `mission_id` and applies the parent's `security_policy` (e.g., "only allow .edu domains").

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        P[Parent Agent] -->|Tool Call + Context| M[MCP Any Middleware]
        M -->|Enriched Call| S[Subagent]
        S -->|Tool Call| M2[MCP Any Middleware]
        M2 -->|Context Validation| U[Upstream Adapter]
    ```
*   **APIs / Interfaces:**
    *   `ContextMiddleware`: Intercepts `tools/call` and `notifications`.
    *   `ContextStore`: SQLite-backed storage for session metadata.
*   **Data Storage/State:**
    *   Context is stored in the "Blackboard" KV store, keyed by `session_id`.

## 5. Alternatives Considered
*   **Manual Injection:** Require developers to pass context as tool arguments. *Rejected: Increases prompt token usage and developer friction.*
*   **Centralized Orchestrator:** Use a single "Brain" that manages all state. *Rejected: Does not scale to decentralized or local-first swarms (OpenClaw).*

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Context headers must be signed to prevent "Context Injection" attacks where a subagent tries to escalate permissions.
*   **Observability:** Traces will include `parent_session_id` to allow visual "Marble Diagrams" of the swarm flow.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
