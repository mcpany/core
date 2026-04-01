# Design Doc: Cross-Node Reasoning Quorum (CNRQ) Adapter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become decentralized and span multiple physical nodes, ensuring reasoning consistency is critical. Current systems suffer from "hallucination variance," where different agents in the same mesh propose conflicting actions due to non-deterministic model outputs.

The CNRQ Adapter provides the infrastructure for decentralized reasoning consensus. It allows disparate nodes to vote on a proposed reasoning path before any state-mutating tool is executed, ensuring mesh-scale consistency.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested voting on reasoning fragments across physical nodes.
    * Neutralize hallucination variance in distributed swarms.
    * Provide a standardized interface for cross-framework consensus (OpenClaw, Claude Code, etc.).
* **Non-Goals:**
    * Synchronous execution of all agent thoughts (consensus is only for high-risk mutations).
    * Centralized control of agent "personalities."

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Ensure 3 agents on different servers agree on a database schema migration before execution.
* **The Happy Path (Tasks):**
    1. Agent A proposes a reasoning path for the migration.
    2. Agent A broadcasts the "Reasoning Proposal" to the CNRQ Hub.
    3. The Hub routes the proposal to Agents B and C on remote nodes.
    4. Agents B and C perform "Speculative Verification" and issue hardware-signed "Consensus Tokens."
    5. The Hub aggregates the tokens and issues a "Quorum Attestation."
    6. Agent A executes the tool only after receiving the attestation.

## 4. Design & Architecture
* **System Flow:**
    [Agent Proposal] -> [CNRQ Hub] -> [P2P Tunnel] -> [Remote Auditor Agents]
    [Remote Agents] -> [Consensus Token] -> [CNRQ Hub] -> [Quorum Attestation] -> [Proposer]
* **APIs / Interfaces:**
    * `POST /consensus/propose`: Submit a reasoning fragment for quorum.
    * `GET /consensus/poll/{proposal_id}`: Retrieve current vote status.
    * `POST /consensus/vote`: Submit a hardware-signed attestation for a proposal.
* **Data Storage/State:**
    * Ephemeral "Proposal Registry" in the Shared KV Store (Blackboard).
    * Hardware signatures stored in a TPM-bound audit log.

## 5. Alternatives Considered
* **Single-Node Supervisor:** Rejected due to being a single point of failure and bottleneck for distributed meshes.
* **Simple Majority (Non-Attested):** Rejected because it is vulnerable to "Sybil" attacks where a compromised agent spawns many subagents to hijack the vote.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All votes must be signed by a TPM/Secure Enclave. Proposals are origin-locked and session-bound.
* **Observability:** Visualized in the "Cognitive Consensus Dashboard" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
