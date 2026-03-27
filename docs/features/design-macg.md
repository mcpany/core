# Design Doc: Multi-Agent Consensus Guard (MACG)

**Status:** Draft
**Created:** 2026-06-20

## 1. Context and Scope
The move toward "Consensus-Driven Tooling" in the agent ecosystem addresses the "Lone Wolf" subagent exploit. However, existing consensus models are vulnerable to "Consensus Poisoning" and semantic drift. MACG provides a hardware-attested, multi-agent quorum system for high-privilege tool execution within MCP Any.

## 2. Goals & Non-Goals
* **Goals:**
  * Enforce hardware-attested quorums for high-privilege tool calls.
  * Mitigate "Consensus Poisoning" by requiring signatures from diverse specialized teammates.
  * Prevent "Recursive Hallucination" in deep swarms through periodic mission-root alignment checks.
* **Non-Goals:**
  * Providing the consensus algorithm itself.
  * Managing the lifecycle of individual subagents.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Privilege Agent Swarm Orchestrator
* **Primary Goal:** Execute a host-level write operation that requires a consensus of three independent security subagents.
* **The Happy Path (Tasks):**
  1. The Orchestrator initiates a high-privilege tool call.
  2. MACG intercepts the call and identifies the required quorum (3/3).
  3. Three specialized security teammates review and sign the tool call fragment using HAIL tokens.
  4. MACG verifies the lineage back to the mission-root TPM.
  5. The tool call is released for execution upon quorum fulfillment.

## 4. Design & Architecture
* **System Flow:** Tool Call -> MACG Interceptor -> Quorum Check -> Teammate Review -> HAIL-Signed Vote -> Quorum Fulfillment -> Execution.
* **APIs / Interfaces:** `POST /v1/consensus/quorum/init`, `POST /v1/consensus/vote`.
* **Data Storage/State:** Quorum state is stored in a mission-bound, hardware-protected "Consensus Blackboard."

## 5. Alternatives Considered
* **Simple Majority Vote**: Rejected as it is vulnerable to consensus poisoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All votes must be HAIL-attested.
* **Observability:** Real-time quorum status visible in the Coordination Handshake Debugger.

## 7. Evolutionary Changelog
* **2026-06-20:** Initial Document Creation.
* **2026-06-21:** Added Deadlock Prevention and CAP Binding logic.
