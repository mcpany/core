# Design Doc: Teammate Reflection Quorum (TRQ) Adapter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agent swarms frequently suffer from "Cognitive Lock," where a single specialist agent's reasoning hallucination is committed to the shared state (e.g., scratchpad or blackboard) and subsequently ingested by the entire team, leading to mission failure. Existing "Self-Correction" mechanisms are individual-centric and fail when the agent's internal model is compromised.

The Teammate Reflection Quorum (TRQ) Adapter evolves the Manifest-Based Reflection (MBR) strategy into a collective event. It mandates that any refined or "self-corrected" reasoning fragment must be attested by a peer "Reviewer Agent" before it can mutate the shared teammate state.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-agent "Reasoning Handshake" for self-correction cycles.
    * Enforce a majority quorum (or peer-review) for any write access to the shared scratchpad after a refinement loop.
    * Detect and block "Refinement Drift" where an agent attempts to bypass mission constraints via self-correction.
    * Standardize the OpenClaw-compliant "Review-before-Commit" signal for Agent Teams.
* **Non-Goals:**
    * Requiring quorums for initial "speculative" reasoning (only for state-mutating refinements).
    * Replacing human-in-the-loop (HITL) for high-stakes tool calls (TRQ is for internal reasoning consistency).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a "Database Specialist" agent's corrected SQL query doesn't accidentally drop a table due to a hallucinated schema change.
* **The Happy Path (Tasks):**
    1. The Database Specialist agent performs a self-correction on its previous SQL plan.
    2. The TRQ Adapter intercepts the commit request to the shared scratchpad.
    3. The Adapter spawns (or selects) a "Reviewer Agent" from the authorized teammate pool.
    4. The Reviewer Agent evaluates the correction against the mission-root manifest and the current schema state.
    5. The Reviewer Agent provides a "Reasoning Attestation" token.
    6. The TRQ Adapter validates the token and allows the mutation to the scratchpad.
    7. Parallel teammates ingest the verified correction.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant A as Specialist Agent
        participant B as TRQ Adapter
        participant C as Reviewer Agent
        participant D as Shared Scratchpad

        A->>B: Propose Refined State (Self-Correction)
        B->>C: Request Peer Review (Reasoning Audit)
        C-->>B: Reasoning Attestation Token (Approved)
        B->>D: Commit Mutation
        B-->>A: Ack
    ```
* **APIs / Interfaces:**
    * `trq.ProposeRefinement(agentID, reasoningFragment, intentID) -> ProposalID`
    * `trq.AttestRefinement(proposalID, reviewResult) -> AttestationToken`
    * `trq.CommitVerified(attestationToken, shardID)`
* **Data Storage/State:**
    * **Refinement Buffer:** Ephemeral storage for pending state mutations awaiting quorum.
    * **Quorum Policy Store:** Defines the number of reviewers and the required confidence scores for different task types.

## 5. Alternatives Considered
* **Single-Agent Self-Reflection (Internal):** Rejected because it is vulnerable to "Cognitive Lock."
* **Centralized "Supervisor" Model:** Rejected because it creates a bottleneck and violates the horizontal mesh principle. TRQ allows for peer-to-peer (P2P) reviews.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Reviewer agents are selected based on their "Reputation-Bound Capability" (RBC) scores.
* **Observability:** Integrated with the "Reflection Quorum Workspace" in the UI for real-time visualization of review cycles.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
