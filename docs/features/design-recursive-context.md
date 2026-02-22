# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
In complex agent swarms, a "Parent" agent often needs to delegate tasks to "Subagents." Currently, this requires the subagent to re-establish context or re-authenticate, leading to friction and high token costs. The Recursive Context Protocol (RCP) provides a standardized way for agents to pass scoped context, including authentication tokens and session state, down the agent hierarchy.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize MCP headers for context propagation.
    * Enable "Zero-Auth" delegation for subagents within the same session.
    * Reduce token usage by allowing subagents to reference parent context.
* **Non-Goals:**
    * Creating a new communication protocol (extends existing MCP JSON-RPC).
    * Managing inter-agent scheduling (handled by orchestrators like CrewAI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars in every prompt.
* **The Happy Path (Tasks):**
    1. Parent agent initiates an MCP session with an `X-MCP-Context-ID`.
    2. Parent agent calls a tool that triggers a subagent.
    3. MCP Any automatically injects the parent's scoped context into the subagent's environment.
    4. Subagent performs the task using the inherited context.
    5. Result is returned to the parent agent with updated context state.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent` -> `MCP Any` -> `Subagent (with inherited headers)` -> `Upstream Service`
* **APIs / Interfaces:**
    * Extended MCP `initialize` request to include `context_inheritance` flags.
    * New middleware to handle `X-MCP-Recursive-Context` headers.
* **Data Storage/State:**
    * Context state is stored in the **Shared Key-Value Store (Blackboard)** and indexed by `Context-ID`.

## 5. Alternatives Considered
* **Manual Context Passing:** Rejected as it increases prompt size and risk of leaking secrets in prompts.
* **Stateful Sessions:** MCP is mostly stateless; RCP adds a layer of "Pseudo-state" via shared storage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Context inheritance is strictly scoped; subagents only receive the minimal set of credentials required for their task.
* **Observability:** Context propagation is traced, allowing users to see the hierarchy of calls in the dashboard.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
