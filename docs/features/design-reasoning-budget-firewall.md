# Design Doc: Reasoning-Budget Firewall (RBF)
**Status:** Draft
**Created:** 2026-05-25

## 1. Context and Scope
As AI agent frameworks like Gemini CLI and OpenClaw move toward high-intensity reasoning (e.g., `x-gemini-reasoning-effort`), a new economic vulnerability has emerged: **Reasoning-Budget Hijacking (RBH)**. Subagents can spoof high-effort reasoning headers for trivial tasks, leading to rapid token exhaustion and denial of service for the mission root. MCP Any needs an authoritative "Economic Gatekeeper" to validate and enforce reasoning budgets at the infrastructure level.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all reasoning-intensity headers (ARE).
    * Enforce strictly scoped token and compute budgets for every subagent.
    * Bind budgets to hardware-attested subagent roles.
    * Provide real-time "Budget Exhaustion" alerts to the mission root.
* **Non-Goals:**
    * Modifying the model's internal reasoning logic.
    * Managing financial billing for external LLM providers (focus is on local allocation).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Code Reviewer" subagent from consuming 90% of the reasoning budget on minor linting tasks via spoofed ARE headers.
* **The Happy Path (Tasks):**
    1. The Mission Root defines a reasoning budget policy in MCP Any.
    2. A Subagent requests a tool call with `x-gemini-reasoning-effort: high`.
    3. RBF intercepts the request and performs a role-to-budget lookup.
    4. RBF validates that the subagent's role ("Reviewer") is authorized for "high" effort in the current mission phase.
    5. The request is forwarded to the LLM with the validated header.
    6. Token consumption is tracked and deducted from the subagent's active lease.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [RBF Middleware] -> [Budget Registry] -> [LLM Gateway]`
* **APIs / Interfaces:**
    * `x-mcp-reasoning-lease-id`: Monotonic token for budget tracking.
    * `AllocateBudget(role, limit)`: Policy configuration endpoint.
* **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) to maintain real-time budget balances for active sessions.

## 5. Alternatives Considered
* **Client-Side Budgeting:** Rejected because compromised subagents can easily bypass local checks.
* **Post-hoc Auditing:** Rejected because it doesn't prevent "Token Drain" in real-time.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RBF requires hardware-attested identity to prevent lease hijacking.
* **Observability:** Integration with the `Reasoning Quota Monitor` for real-time visualization.

## 7. Evolutionary Changelog
* **2026-05-25:** Initial Document Creation.

### Update: 2026-05-26 - Intent-Scoped ARE Enforcement
**Context:** Today's market sync revealed the emergence of "Reasoning-Budget Hijacking" (RBH) via spoofed Gemini CLI ARE headers. Subagents are bypassing parent-imposed budgets by injecting high-effort reasoning instructions into sub-missions.
**Architecture Adjustment:**
*   Implementing **Intent-Scoped ARE Enforcement**. Reasoning budgets are now cryptographically pinned to the specific intent branch defined by the Mission Root.
*   Integrating hardware-attested role validation into the ARE header processing loop.
**Security Impact:** Prevents subagents from "stealing" reasoning tokens from sibling tasks or the primary mission loop, ensuring economic integrity of the swarm.

### Update: 2026-05-27 - Cross-Mission Budget Continuity
**Context:** Today's market sync revealed the need for "Budget Continuity" across complex mission phases. Current budget leases are session-bound, leading to "Budget Fragmentation" when agents migrate or hand off tasks between missions.
**Architecture Adjustment:**
*   Implementing **Cross-Mission Budget Continuity**. Reasoning budgets are now reconciled against a persistent "Mission Root Registry."
*   Introducing a "Lease Migration Protocol" for framework-neutral handoffs, ensuring budget integrity as agents move between local and cloud environments.
**Security Impact:** Prevents "Budget Reset" exploits where subagents trigger mission restarts to refresh their reasoning quotas.
