# Design Doc: UACO v2.1 IPSC Middleware
**Status:** Draft
**Created:** 2026-03-30

## 1. Context and Scope
The release of OpenClaw v2.6 has introduced autonomous self-correction loops, which, while powerful, have led to the "Cognitive Lock" phenomenon where agents enter infinite refinement cycles. UACO v2.1 IPSC (Intent-Preserving Self-Correction) provides a cryptographic and resource-bound framework to govern these loops. MCP Any needs to implement this middleware to ensure swarm stability and prevent resource exhaustion.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement UACO v2.1 IPSC token validation and enforcement.
    * Halt recursive self-correction loops using a configurable "Correction Budget."
    * Mandate "Intent Re-Verification" after exceeding the correction threshold.
    * Provide real-time telemetry on refinement cycles.
* **Non-Goals:**
    * Automatically fixing the logic errors causing the refinement loops.
    * Replacing the high-level Mission Intent (handled by PoI Validator).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm DevOps Engineer
* **Primary Goal:** Prevent a specialized "Code Optimizer" agent from consuming the entire token budget on a single function refinement.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a task with an IPSC-enabled Intent Token.
    2. Subagent performs a tool call, which fails or requires refinement.
    3. Subagent enters a "Self-Correction" phase.
    4. IPSC Middleware intercepts the refinement request, increments the cycle counter, and deducts from the "Correction Budget."
    5. On the 4th consecutive refinement, the IPSC Middleware blocks the request and triggers an "Intent Drift" alert.
    6. The Parent Agent (or user) receives a request to re-verify the mission intent before execution can resume.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [IPSC Middleware] -> [Cycle Check] -> [Budget Deduction] -> [Intent Verification] -> [Upstream]`
* **APIs / Interfaces:**
    * `X-UACO-IPSC` header for carrying correction state and signatures.
    * `IPSCManager` internal interface: `ValidateCorrection(token, context)`, `IncrementCycle(sessionID)`, `ResetBudget(sessionID)`.
* **Data Storage/State:**
    * Ephemeral session state in the Shared KV Store to track cycle counts and remaining budgets.
    * Signed IPSC tokens to prevent state tampering between agents.

## 5. Alternatives Considered
* **Time-Based Timeouts:** Rejected as they don't account for token density or complexity of refinement.
* **Global Rate Limiting:** Rejected as it's too blunt and doesn't distinguish between productive work and "Cognitive Lock" loops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** IPSC tokens must be cryptographically bound to the Parent Intent to prevent subagents from "self-granting" additional budget.
* **Observability:** Integrate with the "Recursive Loop Heatmap" in the UI to visualize refinement hotspots.

## 7. Evolutionary Changelog
* **2026-03-30:** Initial Document Creation.
