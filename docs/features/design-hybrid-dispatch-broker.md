# Design Doc: Hybrid Dispatch Broker (HDB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms evolve from single-instance sessions to multi-node meshes, the ability to route tasks between heterogeneous environments (Local, Private Cloud, Public Cloud) becomes critical. Current "Dispatch" protocols lack unified hardware-attested security and cross-boundary intent persistence.

MCP Any needs to solve this by acting as the authoritative "Dispatch Router," ensuring that tasks delegated from a local Claude Code instance to a cloud-based OpenClaw specialist carry the same cryptographic mission-root authority and security guardrails.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize cross-boundary task routing via hardware-attested "Task Tickets."
    * Maintain "Mission-Root" intent integrity across heterogeneous frameworks.
    * Provide a unified registry for discovering reachable agent nodes (Local/Cloud).
    * Neutralize "Cross-Cloud Shadow Mapping" via aggregated discovery masking.
* **Non-Goals:**
    * Implementation of a new LLM reasoning engine.
    * Managing the underlying container/VM orchestration (left to K8s/Docker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Agent Orchestrator
* **Primary Goal:** Securely delegate a high-trust database migration task to a cloud-resident specialist agent without leaking local environment credentials.
* **The Happy Path (Tasks):**
    1. Local Agent (Claude Code) initiates a "Dispatch" request to MCP Any.
    2. HDB issues a TPM-signed "Task Ticket" containing the scoped intent and mission-root lineage.
    3. HDB queries the CCDN-shielded discovery hub to find an authorized Cloud Specialist.
    4. HDB establishes a secure AMT (Attested Mesh Tunnel) to the target cloud node.
    5. The Cloud Specialist receives the Task Ticket, verifies its lineage, and executes within the attested mission scope.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Local Agent] -->|Dispatch Request| B(Hybrid Dispatch Broker)
        B -->|Issue Task Ticket| C{TPM/Secure Enclave}
        B -->|Lookup| D(CCDN Shielded Registry)
        B -->|Route Task| E[Cloud Specialist Agent]
        E -->|Verify Ticket| B
    ```
* **APIs / Interfaces:**
    * `POST /v1/dispatch`: Submit a task for cross-boundary routing. Returns a Task Ticket ID.
    * `GET /v1/dispatch/tickets/:id`: Retrieve ticket metadata and attestation status.
* **Data Storage/State:**
    * Task Tickets are stored in a hardware-locked SQLite state store, indexed by Mission-Root ID.

## 5. Alternatives Considered
* **Direct Peer-to-Peer Dispatch**: Rejected due to the complexity of managing cross-cloud mTLS and identity mapping between disparate frameworks.
* **Binary State Handoff (BSH) only**: Rejected as it lacks the "Routing" logic needed to find and coordinate with remote agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All Task Tickets are hardware-attested and lineage-bound. Access to target nodes is governed by SCS (Structured Channel Sovereignty).
* **Observability:** Integrated with the Mesh Topology Monitor to visualize cross-boundary task flow and latency.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
