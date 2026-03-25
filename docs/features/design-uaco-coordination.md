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
* **2026-03-20:** Integration with Ephemeral Workspace Trust. Task Cards now include a `required_trust_level` field that must be satisfied by the bidder's ephemeral session token. Added "Bid Quarantining" for agents with high behavioral anomaly scores.
### Update: 2026-03-22 - Agentic SLA Enforcement
**Context:** Multi-agent swarms are increasingly hitting "Recursive Deadlocks" and "Spiral of Death" loops that exhaust token quotas.
**Architecture Adjustment:**
* Introducing a mandatory `resource_contract` field in the UACO Task Card schema.
* Implementing a real-time monitor in the Delegation Engine that preemptively terminates tool chains exceeding the agreed-upon SLA.
**Security Impact:** Prevents resource exhaustion attacks and ensures deterministic reasoning provenance across disparate agent frameworks.

### Update: 2026-03-27 - Consensus-Based Delegation & RID Parental Override
**Context:** Today's research into Claude Code and UACO v1.8 RID reveals the need for multi-agent validation and stronger parent control over sub-delegations.
**Architecture Adjustment:**
* **Consensus Token Requirement**: High-risk task cards can now specify a `consensus_threshold`, requiring multiple signed "Approval Bids" from monitor agents before delegation.
* **RID Parental Override**: Implementing a real-time "Kill Switch" in the Delegation Engine that allows parent agents to revoke sub-delegations if intent drift is detected.
* **Shard-Aware Task Cards**: Task cards now support `required_shards` metadata to facilitate pre-emptive mounting by the Shard Manager.
**Security Impact:** Mitigates "Intent Hijacking" by rogue subagents and provides a distributed safety net for sensitive operations.

### Update: 2026-03-25 - UACO v1.8: Recursive Intent Delegation (RID)
**Context:** Today's leak of UACO v1.8 confirms that "Intent Hijacking" via unauthorized subagent escalation is a critical ecosystem risk.
**Architecture Adjustment:**
* **RID Validation Middleware**: Introducing a new validation stage that parses `delegation_depth` and `mutation_boundaries` from signed UACO v1.8 tokens.
* **Recursive Nonce-Chaining**: Every sub-delegation must now include a hash-chain reference to the parent's intent signature to prevent "Intent Ghosting."
**Security Impact:** Provides a cryptographic ceiling for autonomous subagent actions, ensuring they remain strictly bound to the user's primary mission root.
