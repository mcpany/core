# Design Doc: Distributed Consensus Hub (DCH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As multi-agent swarms scale across distributed multi-node meshes, the risk of "Single-Agent Hallucinations" and inconsistent state transitions increases. Standard single-node HITL or single-framework quorums are insufficient for heterogeneous swarms where agents may be running on disparate physical devices.

The Distributed Consensus Hub (DCH) is required to act as the authoritative "Consensus Broker" for multi-node swarms, utilizing hardware-attested security enclaves to reach agreement on high-risk tool results and state changes.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested, multi-node consensus on tool execution results.
    * Provide a framework-neutral protocol for agents (Claude, OpenClaw, AutoGen) to submit and verify "Attestation Ballots."
    * Support "Speculative Consensus" to minimize reasoning latency while background hardware-bound verification occurs.
    * Integrate with TPM/Secure Enclave for non-repudiable voting.
* **Non-Goals:**
    * Replacing framework-local coordination (like Claude Code teams); it extends it to cross-node scenarios.
    * Managing low-level network replication; it uses the AMT Broker for transport.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Swarm Orchestrator
* **Primary Goal:** Verify the result of a high-risk database migration performed by a remote agent before updating the global state.
* **The Happy Path (Tasks):**
    1. Remote agent (Node B) executes a database migration tool.
    2. Node B submits the result and a "Verification Ballot" to the DCH.
    3. DCH on Node A (Mission Root) receives the ballot and requests attestation from independent "Monitor" and "Auditor" nodes.
    4. Monitor and Auditor nodes verify the result against their local policies and sign an "Attestation Token" using their TPMs.
    5. DCH collects the tokens and, once a quorum is reached, commits the migration result to the global Blackboard.
    6. Mission Root agent resumes reasoning based on the verified state.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        subgraph Node A (Mission Root)
            A[Root Agent] --> B[DCH Broker]
            B --> C[Blackboard Commit]
        end
        subgraph Node B (Worker)
            D[Specialist Agent] --> E[Tool Result]
            E --> F[Ballot Submission]
        end
        subgraph Node C (Auditor)
            G[Auditor Agent] --> H[Ballot Verification]
            H --> I[TPM Signature]
        end
        F --> B
        I --> B
    ```
* **APIs / Interfaces:**
    * `dch.SubmitBallot(resultHash, proofMetadata) -> BallotID`: Submits a result for consensus.
    * `dch.RequestAttestation(ballotID, nodeId) -> Token`: Requests a signature from a peer node.
    * `dch.GetConsensusStatus(ballotID) -> Status`: Returns the current quorum progress.
* **Data Storage/State:**
    * **Consensus Registry:** In-memory store for active ballots and collected signatures.
    * **Trust Matrix:** List of hardware fingerprints for authorized attestation nodes.

## 5. Alternatives Considered
* **Single-Node Quorums:** Rejected because they don't account for physical node failure or local environment tampering.
* **Centralized Database Locking:** Rejected due to performance bottlenecks and "Single Point of Failure" risks in distributed meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All ballots and attestation tokens must be hardware-signed. DCH enforces "Identity-Bound Consensus."
* **Observability:** Integrated with the "Summarization Quorum Hub" in the UI for real-time visualization of consensus progress.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
