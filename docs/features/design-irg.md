# Design Doc: Intent-Resumption Gateway (IRG)
**Status:** Draft | In Review | Approved
**Created:** 2026-06-15

## 1. Context and Scope
As AI agent swarms evolve from linear execution to high-frequency teammate rotation, the "Cognitive Stall" (averaging 400ms-600ms per handoff) has become a critical performance bottleneck. During handoffs, the recipient teammate must ingest the mission-root context and re-attest its authority, leading to reasoning latency. The **Intent-Resumption Gateway (IRG)** aims to solve this by providing "Intent-Resumption Tokens" that allow for sub-100ms teammate rotation within a verified mission scope.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-100ms teammate rotation latency.
    * Provide hardware-attested "Intent-Resumption Tokens" for mission-root continuity.
    * Pre-attest teammate authority before the sub-mission branch fully spawns.
    * Maintain Zero-Trust isolation between teammates while sharing resumption state.
* **Non-Goals:**
    * Does not replace the primary A2A Messaging Hub for general task delegation.
    * Does not manage long-term context storage (handled by the Blackboard).
    * Will not perform full context summarization (delegated to the ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Orchestrator (e.g., Claude Code Team Lead)
* **Primary Goal:** Rapidly rotate between a "Coder" specialist and a "Security Auditor" without incurring cognitive stall.
* **The Happy Path (Tasks):**
    1. The Team Lead agent requests a "Resumption Token" for a specific mission fragment from the IRG.
    2. The IRG validates the request against the hardware-attested mission root.
    3. The IRG issues a TPM-signed "Intent-Resumption Token" containing the pre-verified mission context.
    4. The "Security Auditor" subagent receives the token during its spawn sequence.
    5. The "Security Auditor" resumes reasoning immediately, bypassing the standard 400ms context ingestion phase.

## 4. Design & Architecture
* **System Flow:**
    - `Orchestrator` -> `IRG` (Request Token with Mission Fragment)
    - `IRG` -> `TPM/Secure Enclave` (Sign Token)
    - `IRG` -> `Orchestrator` (Return Token)
    - `Orchestrator` -> `Target Teammate` (Pass Token during Spawn)
    - `Target Teammate` -> `IRG` (Verify Token & Resume Context)
* **APIs / Interfaces:**
    - `POST /irg/v1/token`: Issue a resumption token for a mission fragment.
    - `GET /irg/v1/verify`: Verify and decrypt a resumption token for context resumption.
* **Data Storage/State:**
    - Tokens are ephemeral and session-bound.
    - Context fragments are held in memory-mapped high-speed buffers (BSH compatible).

## 5. Alternatives Considered
* **Persistent Shared Memory:** Rejected due to the risk of "State-Splicing" exploits and the lack of per-handoff attestation.
* **Centralized Context Cache:** Rejected due to centralized bottlenecking and the overhead of repeated remote lookups.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    - Tokens are TPM-signed and hardware-attested.
    - Tokens are single-use and bound to a specific teammate identity and mission fragment.
    - Intent-Resumption Tokens are encrypted to prevent sibling teammates from "Resumption Hijacking."
* **Observability:**
    - Log resumption latency and token issuance success rates.
    - Monitor for "Resumption Drift" where resumed context deviates from the mission root.

## 7. Evolutionary Changelog
* **2026-06-15:** Initial Document Creation.
