# Design Doc: ACR Hub Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms become deeper and more autonomous, the ability to selectively
revoke capabilities in real-time is critical for security and cost control. The
ACR Hub implements the OpenClaw ACR protocol to provide a centralized
revocation
point for all transient capabilities leased to subagents.

## 2. Goals & Non-Goals
* **Goals:**
  - Facilitate sub-millisecond revocation of specific tool capabilities.
  - Synchronize revocation status across heterogeneous framework boundaries.
  - Provide hardware-attested proof of revocation.
* **Non-Goals:**
  - Implementing the logic to *detect* drift (handled by supervisor agents).
  - Managing long-term agent identity (handled by SMI Relay).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Revoke internet access for a specialized subagent once its
  search phase is complete, without stopping its reasoning process.
* **The Happy Path (Tasks):**
  1. Supervisor initiates an ACR Revoke request via the A2A Messaging Hub.
  2. ACR Hub invalidates the subagent's specific "Search" capability token.
  3. ACR Hub broadcasts the revocation to the Transport Gateway.
  4. The subagent's next attempt to use a search tool is blocked at the
  gateway.

## 4. Design & Architecture
* **System Flow:**
  - [Supervisor] -> [ACR Hub] -> [Revocation Manifest] -> [Gateway Enforcement]
* **APIs / Interfaces:**
  - `RevokeCapability(mission_id, agent_id, capability_mask)`
* **Data Storage/State:**
  - Transient "Revocation Bloom Filters" stored in hardware-locked memory.

## 5. Alternatives Considered
- **Session Termination**: Rejected because it destroys context for unaffected
  sibling agents and requires expensive re-planning.
- **Static Scoping**: Rejected because it cannot adapt to dynamic task shifts.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Revocation manifests are TPM-signed to prevent
  bypass.
* **Observability:** Every revocation event is logged to the hardware-attested
  audit stream.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
