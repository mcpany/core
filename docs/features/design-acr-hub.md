# Design Doc: ACR Hub Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the maturation of OpenClaw v3.3.0 and high-frequency reasoning meshes,
the latency between detecting an agent's intent drift and revoking its
capabilities has become a critical vulnerability. Currently, revocation can
take seconds, allowing misaligned agents to execute high-impact tool calls.
The ACR Hub Controller provides sub-millisecond, autonomously reactive
revocation triggered by Active Intent Alignment (AIA) failure signals.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-millisecond capability revocation across all connected
      frameworks (Claude Code, OpenClaw, AutoGen).
    * Integrate directly with AIA heartbeat monitors.
    * Mandate hardware-attested identity verification for all revocation
      signals.
* **Non-Goals:**
    * Providing the reasoning engine for AIA (handled by the AIA Broker).
    * Managing long-term agent banning (handled by the Skill Reputation
      Engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Revoke all tool capabilities from a specialist subagent
  the moment its internal monologue diverges from the mission root.
* **The Happy Path (Tasks):**
    1. Specialist subagent reasoning trace is ingested by AIA Broker.
    2. AIA Broker detects intent drift and issues a "Drift Alert."
    3. ACR Hub receives the Drift Alert and identifies all session-bound
       capabilities.
    4. ACR Hub immediately broadcasts a "Hardware-Locked Revocation" signal
       to the tool bus.
    5. Tool executions are forcefully terminated at the transport level.

## 4. Design & Architecture
* **System Flow:** AIA Broker -> ACR Hub -> Transport-Layer Session Binder
  (TLSB) -> Revocation.
* **APIs / Interfaces:** `POST /api/v1/acr/revoke` requiring hardware-
  attested identity tokens.
* **Data Storage/State:** In-memory, session-bound capability mapping for
  ultra-low latency.

## 5. Alternatives Considered
* **Reactive Polling:** Rejected due to high latency (100ms+ window).
* **Manual Revocation:** Rejected as it cannot match machine-speed swarm
  attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Revocation signals must be cryptographically
  signed by a hardware-attested authority.
* **Observability:** Every revocation event is logged with full reasoning
  lineage for post-mortem analysis.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
