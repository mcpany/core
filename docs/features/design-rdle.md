# Design Doc: Recursive Depth-Limit Enforcer (RDLE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
To neutralize Recursive Shadow Handoffs (CVE-2026-71001), MCP Any must track
the depth of agent delegations. RDLE ensures that subagents cannot bypass
guardrails by initiating unauthorized handoffs beyond a user-defined depth.

## 2. Goals & Non-Goals
* **Goals:**
  - Track recursive delegation depth via hardware-attested headers.
  - Enforce hard-stop limits on handoff chains.
  - Signal depth violations to the supervisor in real-time.
* **Non-Goals:**
  - Restricting horizontal swarm width (handled by the Reasoning-Budget
    Firewall).
  - Monitoring the *content* of the reasoning.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer
* **Primary Goal:** Prevent a coding agent from spawning a subagent that
  spawns another deeper than 3 levels, ensuring human oversight for deep tasks.
* **The Happy Path (Tasks):**
  1. User sets a global `max_delegation_depth=3`.
  2. A Level-2 subagent attempts to spawn a Level-3 agent.
  3. RDLE appends a signed depth-marker (d=3) to the handoff token.
  4. The Level-3 agent attempts a further spawn; RDLE blocks the request and
     raises a P0 alert.

## 4. Design & Architecture
* **System Flow:**
  - [Agent Call] -> [RDLE Middleware] -> [Depth Validation] ->
    [Handoff Approval]
* **APIs / Interfaces:**
  - Header: `X-UAB-Delegation-Depth` (Hardware-attested and monotonic).
* **Data Storage/State:**
  - Stateless validation via cryptographic lineage proofs.

## 5. Alternatives Considered
- **Stateless Token Expire**: Rejected because tokens can be stripped or
  refreshed by compromised subagents; requires hardware-bound depth counters.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses the SMI Relay to verify parentage and lineage
  at each hop.
* **Observability:** Depth-limit hits are visualized in the Swarm Anomaly
  Dashboard and exported to RL telemetry sinks.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
