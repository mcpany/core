# Design Doc: Intent-Bound Token Manager (IBTM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In autonomous AI swarms, resource management (tokens, API credits, compute time) is typically handled at the session or account level. This creates a vulnerability known as "Budget Siphoning," where a specialized or compromised subagent can consume the entire resource quota of a mission for unauthorized reasoning loops or malicious tool calls.

The Intent-Bound Token Manager (IBTM) evolves resource governance by cryptographically tying API quotas and reasoning budgets to specific mission-root intent signatures. By ensuring that resource consumption is non-repudiable and intent-scoped, MCP Any provides hardware-locked economic sovereignty for enterprise agent deployments.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically tie reasoning and token budgets to hardware-attested mission intents.
    * Prevent subagents from consuming resources outside their explicitly assigned intent branch.
    * Provide real-time "Budget Exhaustion" signals to supervisor agents.
    * Support "Mission-Wide Reconciliation" of budgets across heterogeneous agent frameworks.
* **Non-Goals:**
    * Managing the actual payment or billing processing for LLM providers (IBTM is a quota enforcer).
    * Providing token count estimation for non-textual inputs (multimodal budgeting handled by MITS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Resource Manager
* **Primary Goal:** Prevent a "Runaway Reasoning Loop" in a research subagent from consuming a $500 token budget intended for a mission-critical production fix.
* **The Happy Path (Tasks):**
    1. The Manager defines a "Production Fix" mission with a hardware-attested intent signature and a 1M token budget.
    2. A "Research Specialist" subagent is delegated a sub-task with a 50k token limit.
    3. The IBTM issues an "Intent-Bound Token Lease" to the research subagent.
    4. The subagent attempts a high-intensity reasoning loop that would exceed 50k tokens.
    5. The IBTM intercepts the reasoning request via the `ARE Provider`.
    6. The IBTM detects the budget violation and automatically throttles the subagent's reasoning effort.
    7. The IBTM notifies the Mission Root to re-allocate budget or terminate the sub-task.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Mission Root] -> (Attested Intent + Budget) -> [IBTM Broker]
                                                          |
                                                          v
    [Subagent] -> (Reasoning Request + Lease) -> [IBTM Validator]
                                                          |
                                                          v
    [LLM Provider] <--- (Authorized Request) <--- [Quota Ledger]
    ```
* **APIs / Interfaces:**
    * `POST /v1/tokens/lease`: Issues an intent-bound token lease.
    * `GET /v1/tokens/usage/{intent_id}`: Returns real-time budget consumption metrics.
* **Data Storage/State:**
    * Quota ledgers are maintained in the Shared KV Store (Blackboard) using the `HACA` (Hardware-Attested Cost Attribution) standard.
    * Intent signatures are verified via the `SRM Provider`.

## 5. Alternatives Considered
* **Global Session Quotas:** Rejected as they fail to prevent "Insider Threat" subagents from siphoning resources from critical siblings.
* **Time-Based Throttling:** Rejected because reasoning-intensity (tokens) is a more accurate measure of resource impact than wall-clock time.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Token leases are cryptographically bound to the hardware enclave ID. They cannot be transferred between subagents or reused across disparate missions.
* **Observability:** Budget burn rates are visualized in the `Usage Quota Dashboard` and the `Economic Attribution Viewer`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
