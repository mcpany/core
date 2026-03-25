# Design Doc: Mesh-Resident Governance Oracle (MRGO)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As swarms grow in complexity, the centralized Mission Root becomes a single point of failure and a performance bottleneck. MRGO allows a cluster of agents to form a temporary "Governance Quorum" to authorize tool calls without calling back to the root.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable decentralized tool call authorization.
    * Use TPM-signed tokens for peer-to-peer attestation.
    * Minimize latency in high-frequency tool loops.
* **Non-Goals:**
    * Replacing the Mission Root for long-term state audit.
    * Handling cross-swarm governance (limited to local mesh).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Authorize a PR creation tool call using a consensus of 3 subagents.
* **The Happy Path (Tasks):**
    1. Agent A initiates a "Protected Tool Call" (Git PR).
    2. MRGO Adapter broadcasts a "Policy Request" to peer agents.
    3. Peers verify the request against their local mission fragment and return a TPM-signed approval token.
    4. MRGO Adapter aggregates tokens and executes the tool once quorum is reached.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        S[Specialist Teammate] -->|Policy Request| MRGO[MRGO Adapter]
        MRGO -->|Broadcast| Peers[Teammate Mesh]
        Peers -->|TPM-Signed Token| MRGO
        MRGO -->|Verify Quorum| Quorum[Governance Quorum]
        Quorum -->|Valid| Execute[Execute Tool Call]
        Quorum -->|Invalid| Deny[Deny Tool Call]
        Execute -->|Audit Trail| MR[Mission Root]
    ```
* **APIs / Interfaces:** New `/v1/mesh/govern` endpoint for peer-to-peer token exchange.
* **Data Storage/State:** Ephemeral consensus state stored in the Mesh Bus.

## 5. Alternatives Considered
* **Centralized TTL Leases:** Rejected due to lack of resilience during network partitions.
* **Gossip Protocols:** Rejected due to excessive latency for real-time tool authorization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All tokens must be signed by hardware-backed keys (TPM/Secure Enclave).
* **Observability:** Quorum reaching events are logged to the asynchronous RL telemetry stream.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
