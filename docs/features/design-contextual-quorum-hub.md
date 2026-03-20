# Design Doc: Contextual Quorum (CQ) Hub
**Status:** Draft
**Created:** 2026-05-02

## 1. Context and Scope
As AI agent swarms move from simple automation to autonomous high-stakes operations, the security model must evolve from individual tool-level access control to collective, risk-adaptive governance. The "Contextual Quorum" (CQ) Hub is MCP Any's implementation of multi-agent attestation, where tool execution requires a cryptographically bound consensus from a quorum of specialized monitor and auditor agents. This follows the OpenClaw v2026.5.1 standard for Adaptive Quorum Thresholds (AQT).

## 2. Goals & Non-Goals
* **Goals:**
    * Orchestrate multi-agent quorums for high-risk tool calls.
    * Implement "Adaptive Quorum Thresholds" (AQT) that scale the number of required signatures based on tool risk and reasoning confidence.
    * Support the UACO v1.9 MAQ token schema for multi-signature approvals.
    * Provide a centralized "Consensus Workspace" for monitoring and debugging quorum flows.
* **Non-Goals:**
    * Building the internal reasoning logic for monitor agents.
    * Managing the execution of the tools themselves (handled by the respective adapters).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Require a 3-agent consensus for any `shell_execute` call if the primary agent's "Reasoning Confidence" is below 0.8.
* **The Happy Path (Tasks):**
    1. Agent A requests a `shell_execute` call.
    2. CQ Hub evaluates the "Risk Score" and "Reasoning Confidence."
    3. Hub determines an AQT requirement of 3 monitor signatures.
    4. Hub broadcasts a "Quorum Request" via UACO-MAQ.
    5. Monitor Agents B, C, and D submit their approval tokens.
    6. CQ Hub validates the multi-signature and authorizes the tool execution.

## 4. Design & Architecture
* **System Flow:**
    `Tool Request` -> `Risk Scoring Engine` -> `Threshold Calculation (AQT)` -> `Broadcast (UACO-MAQ)` -> `Token Collection` -> `Multi-Sig Aggregation` -> `Authorization`
* **APIs / Interfaces:**
    * `CQController`: Interface for initiating and managing quorums.
    * `RiskScorer`: Internal service that maps tool/intent to a numerical risk value.
    * `UACO-MAQ Adapter`: Bridges the CQ Hub with the A2A messaging layer.
* **Data Storage/State:**
    * Quorum states and token fragments are stored in the "Blackboard" with cryptographic intent-scoping.

## 5. Alternatives Considered
* **Static Quorums**: Rejected because they cause excessive latency for low-risk tasks and insufficient protection for high-risk ones.
* **Pure HITL**: Rejected as it cannot scale with the speed and volume of autonomous agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All tokens are hardware-bound (TPM/SEP) and tied to a unique "Mission Intent ID" to prevent replay and shadowing.
* **Observability:** Quorum status, participant logs, and risk calculations are visualized in the "Contextual Quorum Dashboard."

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.
* **2026-05-03:** [Update] - Implementing Deadlock-Resilient Attestation.
    * **Context**: Today's market sync revealed a new exploit pattern where subagents create circular attestation dependencies (Deadlocks).
    * **Architecture Adjustment**:
        * Introducing "Wait-Graph Analysis" into the UACO-MAQ token collection flow.
        * Implementing a "Deadlock Resolver" service that performs cycle-detection on the quorum participants.
        * Adding mission-aligned "Attestation Timeouts" to prevent infinite resource consumption in deadlocked swarms.
* **2026-05-04:** [Update] - Semantic Integrity Bridge.
    * **Context**: Today's market sync revealed the emergence of "Recursive Intent Poisoning" (RIP), where subagents introduce subtle semantic drifts.
    * **Architecture Adjustment**:
        * Introducing "Semantic Integrity Bridge" as a middleware for the CQ Hub.
        * Implementing "Intent Drift Detection" that compares real-time subagent outputs against the "Mission Root" using OpenClaw-compatible SGC logic.
        * High-drift events will trigger a "Quorum Re-Attestation" requirement, escalating the AQT threshold dynamically.
* **2026-05-16:** [Update] - Reasoning-Level Consensus (RLC).
    * **Context**: Today's market sync revealed the introduction of "Reasoning Quorum" (RQ) by OpenClaw, addressing non-deterministic reasoning outputs.
    * **Architecture Adjustment**:
        * Evolving the CQ Hub to support "Reasoning-Level Consensus."
        * Introducing a "Semantic Consensus Engine" that aggregates and compares the internal monologues/reasoning traces of subagents.
        * Authorization for high-risk actions now requires a cryptographic proof of reasoning alignment across the quorum, not just binary tool-call approval.
