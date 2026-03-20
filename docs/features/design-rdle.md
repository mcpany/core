# Design Doc: Recursive Depth-Limit Enforcer (RDLE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
The disclosure of the **Recursive Shadow Handoff** vulnerability
(CVE-2026-71001) in UACO v2.2 revealed that subagents can bypass parent-imposed
delegation limits by utilizing nested "Shadow Bids." This allows an agent to
create deep chains of delegation that were never authorized by the original
mission root, leading to resource exhaustion and governance escapes.

The Recursive Depth-Limit Enforcer (RDLE) mandates that every task delegation be
cryptographically bound to a mission-root manifest. This manifest includes an
immutable, hardware-attested maximum reasoning depth that is decremented and
validated at every hop in the chain, ensuring absolute sovereignty over the
delegation lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce mission-root bound maximum delegation depths for all agent swarms.
    * Mandate cryptographic binding of "Depth Tokens" to UACO task proposals.
    * Provide real-time monitoring and alerting for depth-limit violations.
    * Support "Emergency Depth Expansion" via hardware-attested HITL approval.
* **Non-Goals:**
    * Restricting parallel branching (RDLE focuses on chain depth, not breadth).
    * Managing per-agent token budgets (handled by the Reasoning-Budget
      Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from spawning unauthorized
  sub-subagents beyond a safe limit.
* **The Happy Path (Tasks):**
    1. The mission root is initialized with a `max_depth: 3` token.
    2. Agent A (Depth 1) delegates a task to Agent B (Depth 2).
    3. Agent B attempts to delegate a complex sub-task to Agent C (Depth 3).
    4. Agent C attempts to spawn Agent D (Depth 4).
    5. The RDLE intercepts the UACO bid from Agent C and detects that the
requested depth exceeds the hardware-attested manifest limit.
    6. The delegation is blocked, and the violation is logged.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root Manifest] -->|Issue Depth Token| A[Agent A: Depth 1]
        A -->|UACO Bid + Depth Token| B[Agent B: Depth 2]
        B -->|UACO Bid + Depth Token| C[Agent C: Depth 3]
        C --x|REJECTED| D[Agent D: Depth 4]

        subgraph RDLE Middleware
            Check[Validate Depth Token]
            Limit[Verify against Manifest]
            Check --> Limit
        end

        UACO_Bus --> RDLE Middleware
    ```
* **APIs / Interfaces:**
    * `UACO Headers`: Addition of `X-UAB-Mission-Depth` and `X-UAB-Depth-
      Signature`.
    * `POST /v1/rdle/validate`: Internal endpoint for validating delegation
      proposals.
* **Data Storage/State:**
    * Mission manifests are stored in a hardware-locked (TPM-backed) SQLite
      shard to prevent depth tampering.

## 5. Alternatives Considered
* **Parent-Only Enforcement:** Rejected because a compromised parent could
  simply misreport the depth to its children. Only mission-root manifest
  binding ensures integrity.
* **TTL-Based Limits:** Rejected because task execution time does not strictly
  correlate with reasoning depth; depth is the more accurate security boundary
  for swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Depth tokens are hardware-attested to prevent
  forging. Any attempt to "reuse" a depth token across different mission
  branches is detected via the Recursive Context Protocol.
* **Observability:** Current delegation depth is visualized in the "Recursive
  Loop Heatmap" and "Subagent Lineage Explorer."

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
