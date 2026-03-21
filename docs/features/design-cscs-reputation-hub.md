# Design Doc: CSCS Reputation Hub
**Status:** Draft
**Created:** 2026-05-04

## 1. Context and Scope
As AI agent swarms proliferate, the security model must transition from local, siloed attestation to a federated, global trust model. The Cross-Swarm Consensus Scoring (CSCS) Reputation Hub is MCP Any's implementation of this federated model, allowing it to ingest trust signals from independent swarms via OpenClaw v2026.5.3. This ensures that tool trust is informed by global behavior, neutralizing "Chain-of-Thought Spoofing" and "Registry Compromise" attacks across the wider ecosystem.

## 2. Goals & Non-Goals
* **Goals:**
    * Ingest and aggregate CSCS reputation signals from verified peer swarms.
    * Provide a "Global Trust Score" for every tool in the discovery registry.
    * Feed reputation data into the `Risk-Adaptive CQ Controller` to dynamically scale quorum requirements.
    * Maintain a cryptographically signed "Audit Trail" of reputation shifts.
* **Non-Goals:**
    * Performing direct behavioral analysis (handled by the Ghost Shell Profiler).
    * Mandating a single global registry (it remains a decentralized mesh).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Swarm Orchestrator
* **Primary Goal:** Automatically increase the quorum threshold for a tool if its CSCS reputation drops below 0.7 across 5+ independent swarms.
* **The Happy Path (Tasks):**
    1. MCP Any's CSCS Hub receives a reputation broadcast from a peer swarm.
    2. The Hub verifies the peer's signature and updates the tool's aggregate CSCS score.
    3. An agent requests the tool.
    4. The `Risk-Adaptive CQ Controller` queries the CSCS Hub.
    5. The Controller identifies the low score and scales the AQT requirement from 2 to 5 signatures.
    6. The user is notified of the "Elevated Quorum" due to global reputation signals.

## 4. Design & Architecture
* **System Flow:**
    `Peer Signal` -> `CSCS Ingestor` -> `Reputation Aggregator` -> `Global Trust Index` -> `CQ Controller Query` -> `Adaptive Thresholding`
* **APIs / Interfaces:**
    * `CSCSReceiver`: gRPC/UACO endpoint for receiving peer broadcasts.
    * `ReputationProvider`: Internal API for querying tool trust scores.
    * `SwarmPeeringManager`: Manages verified trust links between independent MCP Any nodes.
* **Data Storage/State:**
    * Reputation fragments and peer identity keys are stored in the "Mesh-Aware Blackboard" with graph-based reconciliation.

## 5. Alternatives Considered
* **Local-Only Reputation**: Rejected because it cannot protect against "Day Zero" attacks in newly discovered tools.
* **Centralized Trust Authority**: Rejected as it creates a single point of failure and violates the decentralized nature of the Universal Agent Bus.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All peer signals must be signed by a hardware-bound (TPM) identity. We implement "Reputation Poisoning" defense to prevent Sybil-Swarm attacks.
* **Observability:** A "Cross-Swarm Reputation Map" in the UI allows users to visualize which swarms are contributing to a tool's score.

## 7. Evolutionary Changelog
* **2026-05-04:** Initial Document Creation.
