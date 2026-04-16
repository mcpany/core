# Design Doc: Reflection-Budget Controller (RBC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of "Teammate Reflection" in horizontal Agent Teams (Claude Code) has led to a new stability failure: **Reflection Storms**. Parallel teammates can enter infinite loops of mutual self-correction while trying to reconcile conflicting state on a shared scratchpad. MCP Any needs an RBC to govern the "Self-Correction" lifecycle, enforcing hardware-attested turn limits on reasoning refinements to prevent token exhaustion and cognitive deadlock.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and track "Reflection" or "Self-Correction" reasoning turns.
    * Enforce strictly scoped turn budgets for every subagent mission branch.
    * Trigger automated "Conflict Resolution" or "Supervisor Escalation" when budgets are depleted.
    * Neutralize "Token Storms" by halting recursive refinement loops.
* **Non-Goals:**
    * Determining the *quality* of the reflection (focus is on quantity/budget).
    * Modifying the agent's internal correction logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Prevent two agents from spending $50 in tokens "correcting" each other's minor formatting preferences in a 10-minute infinite loop.
* **The Happy Path (Tasks):**
    1. Mission Root sets a `max_reflection_turns: 3` policy.
    2. Subagent A attempts its 4th "self-correction" turn on the Blackboard.
    3. RBC intercepts the write request and detects budget exhaustion.
    4. RBC blocks the turn and signals the `Mission-Root Conflict Resolver (MRCR)`.
    5. The MRCR applies a priority-weighted fix, breaking the loop.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Reflection Turn] -> [RBC Middleware] -> [Turn Counter (Blackboard)] -> [Arbiter/Escalation]`
* **APIs / Interfaces:**
    * `x-mcp-reflection-depth`: Monotonic header tracking the refinement turn count.
    * `OnBudgetExhausted(agent_id, mission_id)`: Escalation trigger.
* **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) to maintain monotonic turn counters per intent branch.

## 5. Alternatives Considered
* **Timeout-based Halting:** Rejected because high-latency but valid tasks might be prematurely killed.
* **Manual Intervention only:** Rejected as it doesn't scale for autonomous machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RBC budgets are TPM-signed; agents cannot "reset" their turn counter by spawning a shadow subagent.
* **Observability:** Integration with the `IPSC Correction Monitor` for real-time visualization of refinement depth.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
