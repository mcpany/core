# Design Doc: UACO-Native Coordination Middleware
**Status:** Draft
**Created:** 2026-03-19

## 1. Context and Scope
As AI agent ecosystems transition from solitary tools to multi-agent swarms, the bottleneck has shifted from "Tool Execution" to "Task Coordination." The Universal Agent Coordination Protocol (UACO) provides a standardized framework for agents to negotiate, bid on, and delegate tasks. MCP Any must implement a native UACO coordination layer to facilitate reliable, framework-neutral handoffs (e.g., between OpenClaw and AutoGen).

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware layer that parses and validates UACO task negotiation messages.
    * Provide a standardized "Bidding" interface for agents to express capabilities and resource requirements.
    * Ensure "Stateful Handoffs" by cryptographically binding execution context to UACO delegation requests.
    * Integration with the Shared KV Store (Blackboard) for coordinating task state.
* **Non-Goals:**
    * Implementing the low-level transport for UACO (handled by the A2A Bridge).
    * Providing the "intelligence" for bidding (handled by the individual agents).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent System Architect
* **Primary Goal:** Coordinate a complex multi-step research task between a Specialized Researcher (OpenClaw) and a Writer (AutoGen).
* **The Happy Path (Tasks):**
    1. Parent Agent (Researcher) creates a UACO "Task Card" for a writing sub-task.
    2. MCP Any broadcasts the Task Card to available subagents.
    3. Writer Agent (AutoGen) submits a "Bid" via UACO, specifying its token availability.
    4. MCP Any validates the bid and facilitates the "Stateful Handoff," transferring necessary context to the Writer.
    5. The Writer completes the task and returns the result to the Researcher via the UACO completion schema.

## 4. Design & Architecture
* **System Flow:**
    `Task Card` -> `UACO Middleware` -> `Discovery Service` -> `Bidding Loop` -> `Delegation Engine` -> `Stateful Handoff`
* **APIs / Interfaces:**
    * `UACOCoordinator` Interface: `Negotiate(task *UACOTask) (*UACOBid, error)`
    * Internal Message Bus: Support for `UACO.Negotiate`, `UACO.Bid`, `UACO.Delegate`, `UACO.Complete`.
* **Data Storage/State:**
    * Active task cards and bids are stored in the Shared KV Store (Blackboard) with "Swarm-Scoped" isolation.

## 5. Alternatives Considered
* **Framework-Specific Handoffs**: Rejected as it creates vendor lock-in and prevents interoperability between OpenClaw and other frameworks.
* **Pure MCP Tool Calls for Delegation**: Rejected because tool calls lack the richness required for complex negotiation (e.g., bidding, resource constraints).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All UACO messages must be signed using the "Signed Context Chain" to prevent identity spoofing during delegation.
* **Observability:** UACO negotiation steps are logged to the "Agent Chain Tracer (A2A)" for debugging handoff failures.

## 7. Evolutionary Changelog
* **2026-03-19:** Initial Document Creation.
