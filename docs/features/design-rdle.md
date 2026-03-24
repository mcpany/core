# Design Doc: Recursive Depth-Limit Enforcer (RDLE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
To prevent Recursive Shadow Handoffs (CVE-2026-71001), MCP Any needs a
mechanism to track the depth of delegation chains.

## 2. Goals & Non-Goals
* **Goals:**
  - Track recursive handoff depth.
  - Enforce hard limits on delegation loops.
* **Non-Goals:**
  - Monitoring reasoning quality.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer
* **Primary Goal:** Prevent subagents from delegating tasks more than 3 levels
  deep.
* **Happy Path:**
  1. Root agent initiates a task.
  2. RDLE appends a signed depth-marker (d=0).
  3. Subagent attempts level 4 delegation.
  4. RDLE blocks the request and signals a guardrail violation.

## 4. Design & Architecture
RDLE utilizes hardware-attested headers to ensure markers cannot be stripped.

## 5. Alternatives Considered
* **Stateless Token Expire:** Rejected as it doesn't prevent horizontal branch
  explosion.

## 6. Cross-Cutting Concerns
* **Observability:** Depth violations are exported to RL telemetry sinks.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
