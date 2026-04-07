# Design Doc: Active Teammate Reflection (ATR) Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of horizontal "Agent Teams" (e.g., Claude Code v3.3.0), teammates coordinate via a shared mailbox. However, simply identifying an agent is no longer sufficient to ensure swarm stability. A specialist agent might be cryptographically valid but semantically misaligned, leading to "Cognitive Drift" or redundant task execution.

The ATR Arbiter mandates that agents provide a "Reasoning Justification" before they are allowed to claim a task. This moves security and coordination from "Who are you?" to "What are you thinking, and does it align with the mission?"

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate hardware-attested reasoning justifications for all task claims in horizontal swarms.
    * Provide a peer-review mechanism where the mesh can reach consensus on an agent's proposed cognitive path.
    * Reduce "Cognitive Stall" by utilizing optimistic reflection processing for low-risk tasks.
* **Non-Goals:**
    * The arbiter does not rewrite the agent's reasoning; it only validates it against the mission-root manifest.
    * It does not replace the primary Policy Firewall.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a "Database Specialist" subagent only modifies the production schema if its internal reasoning shows it has correctly identified the impact.
* **The Happy Path (Tasks):**
    1. A "Schema Migration" task is posted to the teammate mailbox.
    2. The Database Specialist agent attempts to claim the task.
    3. The ATR Arbiter intercedes, requesting a "Reasoning Justification."
    4. The agent broadcasts a reasoning fragment: "I will first run a dry-run to verify the foreign key constraints..."
    5. The ATR Arbiter (backed by the Mission-Root agent) validates this justification against the `Mission Manifest`.
    6. Consensus is reached, and the agent is granted a task-claim token.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Teammate->>Mailbox: Attempt Task Claim
        Mailbox-->>ATR Arbiter: Intercept Claim
        ATR Arbiter->>Teammate: Request Justification
        Teammate-->>ATR Arbiter: Broadcast Reason Fragment (TPM-Signed)
        ATR Arbiter->>Mesh: Request Peer Attestation
        Mesh-->>ATR Arbiter: Consensus Reached
        ATR Arbiter->>Teammate: Issue Claim Token
    ```
* **APIs / Interfaces:**
    * `ATR.RequestJustification(agentID: NHI, taskID: GUID): challenge`
    * `ATR.ValidateReflection(reason: ReasoningFragment): ConsensusResult`
* **Data Storage/State:**
    * Justifications are stored in the `Shared KV Store (Blackboard)` under the `reflection/` namespace.
    * Conflict-Free Replicated Data Types (CRDTs) are used to manage consensus votes.

## 5. Alternatives Considered
* **Implicit Trust:** Rejected due to the risk of specialized agents diverging from the parent mission root during long-running sessions.
* **Synchronous Global Lock:** Rejected to prevent the "Cognitive Stall" observed in legacy team coordination models.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ATR ensures that even "internal" teammates must justify their actions semantically.
* **Observability:** All justifications and consensus logs are visible in the `Reasoning Alignment Visualizer`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
