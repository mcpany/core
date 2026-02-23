# Design Doc: Recursive Context Protocol (RCP)
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
Hierarchical agent swarms (like OpenClaw or CrewAI) often lose critical context when a parent agent delegates a task to a child. This leads to redundant configuration, loss of authorization, and hallucinations. RCP standardizes how context is inherited down the agent tree.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize context headers for MCP requests.
    * Enable parent agents to pass restricted session tokens to subagents.
    * Provide a mechanism for subagents to report results back up the chain with preserved trace IDs.
* **Non-Goals:**
    * Automatically synchronizing large data blobs between agents (use the Blackboard for that).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a session with an "Authorization Envelope".
    2. Parent agent calls a subagent (via an MCP tool).
    3. MCP Any automatically injects the RCP headers into the subagent's environment.
    4. The subagent uses the inherited context to call an upstream tool (e.g., GitHub API) without needing its own API key.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent -> MCP Any (Tool Call) -> Subagent (with RCP Headers) -> MCP Any -> Upstream`
* **APIs / Interfaces:**
    * Header: `X-MCP-Context-ID`
    * Header: `X-MCP-Inherited-Auth`
* **Data Storage/State:**
    * Context state is transient and bound to the session, managed by the Service Registry.

## 5. Alternatives Considered
* **Manual Env Var Injection:** Rejected because it is insecure and platform-specific.
* **Global Auth Store:** Rejected because it doesn't allow for granular, session-bound permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RCP uses "Least Privilege" for inheritance. A subagent only receives the context it needs, never the parent's master key.
* **Observability:** Trace IDs are propagated through RCP headers, enabling end-to-end visualization of swarm activity.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
