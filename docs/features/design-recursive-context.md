# Design Doc: Recursive Context Protocol

**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
In multi-agent swarms (e.g., OpenClaw), a parent agent often spawns child agents to handle sub-tasks. Currently, there is no standardized way to pass security context, session IDs, or resource limits from the parent to the child. This leads to "context fragmentation" and security vulnerabilities. The Recursive Context Protocol aims to standardize how this metadata is propagated across agent boundaries.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Define a set of standardized MCP headers for context propagation (`X-MCP-Parent-ID`, `X-MCP-Session-Trace`, etc.).
    *   Enable child agents to automatically inherit parent permissions (with optional attenuation).
    *   Support tracking of "Agent Lineage" for audit purposes.
*   **Non-Goals:**
    *   Standardizing the agent communication protocol itself (e.g., how agents send messages).
    *   Handling data synchronization between agents (handled by Shared KV Store).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator initializes a session with a `trace_id`.
    2.  Agent A (Parent) calls a tool to spawn Agent B (Child).
    3.  MCP Any intercepts the call and injects the `trace_id` and Parent's `auth_scope` into the child's environment.
    4.  Agent B executes a tool call.
    5.  The Policy Firewall sees that Agent B is a child of Agent A and allows the call based on inherited permissions.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Client -->|Session Context| AgentA[Agent A (Parent)]
        AgentA -->|Spawn Request| MCP[MCP Any]
        MCP -->|Inject Metadata| AgentB[Agent B (Child)]
        AgentB -->|Tool Call + TraceID| MCP
        MCP -->|Validate Lineage| Firewall[Policy Firewall]
    ```
*   **APIs / Interfaces:**
    *   New standardized headers in JSON-RPC `meta` field.
*   **Data Storage/State:**
    *   Lineage tree is kept in-memory for the duration of the session.

## 5. Alternatives Considered
*   **Manual Context Passing:** Too error-prone and developer-intensive.
*   **Global Shared Context:** Violates Zero Trust principles by giving all agents the same permissions.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Child agents should always have equal or *less* permission than their parents (attenuation).
*   **Observability:** Provide a "Lineage Explorer" in the UI to visualize agent hierarchies.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
