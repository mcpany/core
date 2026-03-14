# Design Doc: Speculative Auction Broker (SAB)
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
As AI agent swarms grow in depth and complexity, the overhead of coordinating and delegating tasks becomes a significant bottleneck, often referred to as "Negotiation Exhaustion." Sequential task bidding and manual orchestration cannot keep pace with high-frequency reasoning loops, especially those utilizing speculative execution patterns (e.g., Gemini-style "Intent Probability" branching).

The Speculative Auction Broker (SAB) is a high-speed negotiation hub designed to arbitrate between multiple competing agentic branches. By utilizing hardware-accelerated negotiation (HAN) and intent-aware bidding, MCP Any can minimize the "Cognitive Stall" in deep swarms and ensure that resources are allocated to the most probable successful reasoning path.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a low-latency (sub-millisecond) bus for agent capability bidding.
    * Support "Intent Probability" scoring to allow orchestrators to prioritize specific reasoning branches.
    * Integrate with hardware-bound (TPM/Secure Enclave) attestation to prevent "Bid Shadowing" attacks.
    * Facilitate state-isolated speculative execution by managing temporary resource locks.
* **Non-Goals:**
    * Executing the LLM reasoning logic itself.
    * Managing persistent agent memory (handled by the Shared Blackboard).
    * Providing a general-purpose agent framework (SAB is an infrastructure layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Rapidly delegate sub-tasks to the most capable subagents across multiple speculative branches without increasing mission latency.
* **The Happy Path (Tasks):**
    1. The Parent Agent broadcasts a Task Card containing multiple speculative intent paths to the SAB.
    2. Available Subagents evaluate the Task Card and submit BSH-encoded bids, including Capability Claims and probability-weighted cost estimates.
    3. The SAB (utilizing the HAN Broker) arbitrates the bids in a hardware-isolated environment.
    4. The SAB notifies the winning subagents and provides them with ephemeral, intent-bound tokens for Blackboard access.
    5. The Swarm proceeds with parallel execution, and the SAB monitors resource consumption against the bid SLA.

## 4. Design & Architecture
* **System Flow:**
    `[Orchestrator] -> [UACO Task Card] -> [SAB (HAN)] -> [Winning Subagents] -> [Ephemeral Blackboard Access]`
* **APIs / Interfaces:**
    * `uab/auction/submit`: Endpoint for parents to post task cards.
    * `uab/auction/bid`: Endpoint for subagents to submit cryptographically signed bids.
    * `uab/auction/status`: Real-time stream of auction results and intent alignment scores.
* **Data Storage/State:**
    * **Auction State**: Held in zero-copy, memory-mapped regions to minimize latency.
    * **Bid Ledger**: An immutable, session-bound audit trail for post-mission optimization.

## 5. Alternatives Considered
* **Sequential Negotiation**: Rejected due to 100ms+ overhead per handoff, which causes "Cognitive Stall" in deep swarms.
* **Static Capability Mapping**: Rejected because agent availability and "Reasoning Health" are dynamic and cannot be predicted at startup.
* **Centralized LLM Orchestrator**: Rejected due to excessive token costs and the risk of the orchestrator becoming a single point of failure (and a bottleneck).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All bids must be signed by a hardware-bound identity. The SAB performs "Pre-Flight Profiling" to ensure bids align with the agent's historical behavioral baseline.
* **Observability:** Provides high-resolution telemetry on negotiation latency, bid-density, and "Auction Health" (win/loss ratios).

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
