# Design Doc: ASH Consensus Broker
**Status:** Draft
**Created:** 2026-04-20

## 1. Context and Scope
With the introduction of Autonomous Self-Healing (ASH) in OpenClaw v2.8, agent swarms now have the capability to detect and correct "Cognitive Drift" autonomously. However, this creates a new risk: a single compromised or "hallucinating" agent could trigger a malicious re-alignment. The ASH Consensus Broker provides a decentralized governance layer within MCP Any that requires a multi-agent quorum to authorize state rollbacks and reasoning path re-alignments.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Consensus-Based Re-alignment" protocol for multi-agent swarms.
    * Provide a secure voting mechanism where agents can attest to the sanity of a reasoning path.
    * Integrate with the Blackboard Versioning Hub to execute atomic state rollbacks upon consensus.
    * Minimize latency in the voting cycle using UACO v2.5 Trust Leases.
* **Non-Goals:**
    * Providing the heuristic logic for detecting drift (this remains the responsibility of the agents/monitors).
    * Enforcing consensus on low-risk, non-state-mutating agent actions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Monitor
* **Primary Goal:** Detect a "Cognitive Loop" in a deep subagent chain and orchestrate a swarm-wide rollback to a known-good state.
* **The Happy Path (Tasks):**
    1. A monitor agent detects a deviation from the Root Mission Intent and issues a "Drift Alert" to MCP Any.
    2. The ASH Consensus Broker initiates a "Re-alignment Vote" and notifies all relevant agents in the swarm.
    3. Agents review the proposed "Sanity Checkpoint" and submit cryptographically signed approval tokens.
    4. Once the Multi-Agent Quorum (MAQ) threshold is reached, the Broker triggers a rollback in the Blackboard Versioning Hub.
    5. The swarm resumes execution from the last verified sane state, with the offending intent branch pruned.

## 4. Design & Architecture
* **System Flow:**
    `[Drift Alert] -> [ASH Consensus Broker] -> [Quorum Collection] -> [Blackboard Rollback] -> [Swarm Resume]`
* **APIs / Interfaces:**
    * `/v1/ash/vote`: Endpoint for agents to submit consensus tokens.
    * `/v1/ash/status`: Real-time status of active re-alignment cycles and quorum progress.
    * `ASH Task Object`: Extension of the A2A task card containing state hashes and rollback pointers.
* **Data Storage/State:**
    * Relies on the Blackboard Versioning Hub for immutable state snapshots.
    * Uses the UAB Reputation Engine to weigh agent votes based on their historical reliability.

## 5. Alternatives Considered
* **Centralized Orchestrator Rollback:** Rejected as it creates a single point of failure and doesn't account for the specialized knowledge of subagents who may be better positioned to judge reasoning sanity.
* **Agent-Local Rollbacks:** Rejected as it leads to "State Inconsistency" where one agent rolls back but the global Blackboard remains in a corrupted state.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Votes are only accepted from agents with a valid, origin-locked session token. The MAQ threshold is dynamically adjusted based on the swarm's collective reputation.
* **Observability:** Integrated with the "ASH Rollback Manager" UI to provide a visual breakdown of why a rollback occurred and which agents supported it.

## 7. Evolutionary Changelog
* **2026-04-20:** Initial Document Creation.
* **2026-04-21:** Update: A2UI Integration for Interactive Consensus.
    * **Context**: Research into OpenClaw's "Adaptive Reasoning" reveals that ASH cycles often require human-in-the-loop (HITL) intervention for ambiguous drift.
    * **Architecture Adjustment**: Integrating the ASH Consensus Broker with the A2UI Native Gateway to surface "Consensus Conflict Dashboards" when autonomous quorum fails.
