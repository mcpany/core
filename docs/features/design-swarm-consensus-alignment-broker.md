# Design Doc: Swarm Consensus Alignment Broker
**Status:** Draft
**Created:** 2026-04-18

## 1. Context and Scope
In deep agent swarms, specialized subagents can suffer from "Consensus Drift," where their internal monologue and local state diverge from the parent agent's verified mission intent. This leads to hallucinations, conflicting actions, and mission failure. MCP Any needs a centralized "Alignment Broker" that periodically reconciles subagent states against the authoritative mission monologue to ensure swarm-wide coherence.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized interface for agents to report their "Internal Monologue" and state snapshots.
    * Perform semantic comparison between subagent state and the "Root Mission Intent."
    * Orchestrate state reconciliation (correction) using Multi-Agent Quorum (MAQ) attestation.
    * Halt divergent sub-intents before they commit permanent changes to the Blackboard.
* **Non-Goals:**
    * Replacing the Agent Framework's reasoning engine (the Broker acts as a monitor/auditor).
    * Storing full conversation histories (it only manages state summaries and "proofs of alignment").

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Ensure a specialized "Database Optimizer" subagent doesn't drift into "File Deletion" during a performance-tuning mission.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a "Performance Tuning" mission and registers it with the Broker.
    2. Subagent begins optimization tasks and periodically sends "Alignment Heartbeats" to the Broker.
    3. The Broker detects that the subagent's monologue mentions `rm -rf` on logs, which diverges from the "Optimization" intent.
    4. The Broker triggers a Consensus Check, requesting attestation from the "Security Monitor" agent.
    5. The Security Monitor identifies the drift; the Broker blocks the `rm` tool call and instructs the subagent to rollback to the last verified alignment checkpoint.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent Monologue] -> [Alignment Scorer] -> [Divergence Detector] -> [MAQ Consensus Gate] -> [Correction Signal]`
* **APIs / Interfaces:**
    * `AlignmentBroker`: `ReportMonologue(sessionID string, monologue string) (AlignmentStatus, error)`
    * `AlignmentBroker`: `ReconcileState(sessionID string) (RollbackProof, error)`
* **Data Storage/State:**
    * Stores "Alignment Checkpoints" (State Hashes) in the Blackboard.
    * Uses a TTL-cached "Monologue Buffer" for real-time drift analysis.

## 5. Alternatives Considered
* **Manual HITL for Every Step**: Rejected due to latency and human-bottlenecking.
* **Parent-Only Monitoring**: Rejected because parents in deep chains can also drift or become overwhelmed by subagent telemetry.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Alignment proofs must be signed using hardware-bound keys to prevent "Drift Spoofing."
* **Observability:** Consensus events and drift scores are exported to the "Swarm Truth Explorer" UI.

## 7. Evolutionary Changelog
* **2026-04-18:** Initial Document Creation.
