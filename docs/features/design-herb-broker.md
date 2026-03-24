# Design Doc: Hardware-Enforced Resource Broker
**Status:** Draft
**Created:** 2026-05-18

## 1. Context and Scope
Autonomous agent swarms, especially those with recursive spawning capabilities, are vulnerable to "Resource Hijacking" and "Economic Exhaustion" attacks. A malicious or malfunctioning subagent can spawn infinite siblings or perform compute-heavy tasks that exhaust the entire swarm's token and compute budget before the user or parent agent can intervene.

The Hardware-Enforced Resource Broker (HERB) implements the Task-Bound Resource Isolation (TBRI) standard. It allows Parent agents to allocate hardware-bound "Resource Leases" to specific intent branches. These leases are physically enforced by MCP Any via the local HSM/TPM, ensuring that subagents cannot exceed their allocated quotas regardless of their logical reasoning state.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement deterministic compute and token budgeting for autonomous swarms.
    * Use TPM-bound "Resource Leases" to provide physical enforcement of quotas.
    * Provide real-time telemetry on budget consumption per intent branch.
    * Support "Emergency Budget Revocation" for compromised subagents.
* **Non-Goals:**
    * Managing financial payments/billing (HERB manages the *consumption* of already-allocated resources).
    * Optimizing agent reasoning (HERB is a security and governance gate).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a recursive code-generation loop from consuming $1,000 in tokens in a single session.
* **The Happy Path (Tasks):**
    1. The Architect defines a "Recursive Branch Limit" policy (e.g., Max 500k tokens per branch).
    2. A Parent Agent spawns a Code Specialist and allocates a 100k token Resource Lease.
    3. MCP Any HERB binds this lease to the Code Specialist's hardware-attested ID and mission-root.
    4. The Code Specialist attempts to spawn 5 subagents.
    5. HERB verifies that the total budget for the branch (including the new siblings) does not exceed the Parent's lease.
    6. If a subagent enters an infinite loop, HERB physically blocks subsequent tool calls and LLM requests once the 100k limit is reached.
    7. The Parent Agent receives a "Lease Exhausted" signal and can decide whether to terminate the branch or allocate more resources.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] -> [Request Lease (TBRI)] -> [HERB Broker] -> [TPM-Locked Quota] -> [Subagent Request] -> [Quota Check] -> [Execution]
    1. HERB receives lease request from Parent.
    2. HERB calls TPM to provision a secure counter bound to the Subagent ID.
    3. Every tool call or LLM request mediated by MCP Any decrements the TPM-locked counter.
    4. If counter == 0, execution is force-halted at the gateway level.
* **APIs / Interfaces:**
    * `CreateResourceLease(parent_token, subagent_id, quotas) -> lease_id`
    * `CheckLease(subagent_id, consumption_metrics) -> bool`
    * `RevokeLease(parent_token, lease_id)`
* **Data Storage/State:**
    * Quota counters are stored in volatile TPM/HSM slots for maximum security.
    * Consumption logs are persisted in the Recursive Attestation Ledger (RAR).

## 5. Alternatives Considered
* **Software-based Monitoring:** Rejected; malicious subagents can often disable or bypass software monitoring logic if they achieve high privilege.
* **Centralized API Throttling:** Rejected; too coarse-grained. Swarms need per-intent isolation, not just per-user rate limiting.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Resource limits are physically immutable once signed by the Parent.
* **Observability:** "Budget Near-Exhaustion" alerts are surfaced to the UI and Parent agent.

## 7. Evolutionary Changelog
* **2026-05-18:** Initial Document Creation.
