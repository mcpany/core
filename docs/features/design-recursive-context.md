# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-19

## 1. Context and Scope
In multi-agent swarms (e.g., CrewAI, AutoGen), a parent agent often spawns subagents to perform specialized tasks. Currently, passing configuration, authentication, and session context down this chain is manual and error-prone, leading to "context fragmentation".

The **Recursive Context Protocol** aims to standardize how MCP Any propagates context through the agent hierarchy, ensuring that subagents inherit necessary permissions and state without manual re-configuration.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize a set of MCP header extensions for context propagation.
    * Implement "Context Inheritance" rules in the core server.
    * Support "State Passing" between parent and child agents via the shared KV store.
* **Non-Goals:**
    * Automatically resolving conflicting configurations between parent and child.
    * Managing the lifecycle of subagent processes (handled by the orchestrator).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context (e.g., an encrypted session token) between 3 agents without exposing it to the LLM's prompt.
* **The Happy Path (Tasks):**
    1. Parent Agent receives a tool call with a `X-MCP-Context-ID` header.
    2. Parent Agent calls a subagent tool via MCP Any.
    3. MCP Any detects the subagent call and automatically injects the `X-MCP-Context-ID` and associated metadata into the subagent's environment.
    4. Subagent executes its tool using the inherited context.
    5. The entire chain is linked in the Trace Visualizer via the shared Context ID.

## 4. Design & Architecture
* **System Flow:**
    - MCP Any acts as a "Context Hub".
    - When a tool call is marked as `type: subagent`, MCP Any wraps the call with the current session's metadata.
* **APIs / Interfaces:**
    - New Header: `X-MCP-Recursive-Depth` to prevent infinite loops.
    - New Header: `X-MCP-Context-Inherit: all | none | keys,...`

## 5. Alternatives Considered
* **Manual Passing:** Rejected due to high developer friction and security risks of LLMs seeing raw tokens.

## 6. Cross-Cutting Concerns
* **Security:** Preventing "Context Escalation" where a subagent gets more permissions than its parent.
* **Observability:** Waterfall charts should show the recursive depth of calls.

## 7. Evolutionary Changelog
* **2026-02-19:** Initial Document Creation.
