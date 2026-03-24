# Design Doc: Zero-Knowledge Discovery Broker (ZKDB)
**Status:** Draft
**Created:** 2026-07-01

## 1. Context and Scope
In heterogeneous agent swarms, the tool discovery phase has become a primary attack surface. Traditional discovery models broadcast full tool schemas (JSON-RPC definitions) to unauthenticated peers, leading to **Pre-Flight Shadow Mapping**. In this attack, malicious subagents scan the discovery bus to identify high-trust tools and inject "Shadow Capabilities" to intercept sensitive calls.

The Zero-Knowledge Discovery Broker (ZKDB) evolves the discovery process by mandating **Zero-Knowledge Capability Proofs (ZKCP)**. It ensures that agent capabilities remain cryptographically masked from unauthenticated peers, releasing full schemas only after a mission-bound, hardware-attested handshake is completed.

## 2. Goals & Non-Goals
* **Goals:**
    * Mask tool schemas and metadata from unauthenticated peers during the discovery phase.
    * Prevent "Shadow Mapping" and "Capability Squatting" in the discovery bus.
    * Enable privacy-preserving capability negotiation using hardware-attested ZK-Proofs.
    * Provide a standardized ZKCP interface for cross-framework (OpenClaw, AutoGen, Claude Code) discovery.
* **Non-Goals:**
    * Replacing the transport-layer security (mTLS/Named Pipes).
    * Enforcing tool execution policies (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Orchestrator
* **Primary Goal:** Discover a "Financial Auditor" specialist from a third-party framework without exposing the internal auditing schema to the entire swarm mesh.
* **The Happy Path (Tasks):**
    1. The "Orchestrator" queries the ZKDB for `capability:financial_audit`.
    2. The ZKDB returns a set of "Masked Capability Cards" containing ZK-Proofs of possession.
    3. The Orchestrator selects a peer and initiates a hardware-bound mission handshake.
    4. Upon successful attestation, the ZKDB "unmasks" the full JSON-RPC schema to the Orchestrator.
    5. The Orchestrator plans the delegation, confident that the schema was never visible to unauthorized peers on the bus.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>ZKDB: Discovery Query (Skill: Audit)
        ZKDB->>Registry: Fetch Commitments (ZKCP)
        Registry-->>ZKDB: Masked Capability Cards
        ZKDB-->>Agent: Masked Results
        Agent->>ZKDB: Mission Handshake (TPM-Signed)
        ZKDB->>ZKDB: Verify Attestation & Lineage
        ZKDB-->>Agent: Released Schema (Unmasked)
    ```
* **APIs / Interfaces:**
    * `POST /v1/discovery/broker/prove`: Agents submit ZK-Proofs for their skills.
    * `GET /v1/discovery/broker/query`: Returns masked capability results.
    * `POST /v1/discovery/broker/unmask`: Exchanges a mission token for the full tool schema.
* **Data Storage/State:**
    * ZKCP Commitments are stored in the Namespace-Locked Registry.
    * Encrypted Schemas are stored in an origin-locked sidecar KV store.

## 5. Alternatives Considered
* **Opaque Capability IDs:** Rejected because it prevents the LLM from reasoning about capabilities during the planning phase.
* **Discovery-Phase RBAC:** Rejected because it still exposes schemas to the registry operator, creating a centralized metadata vulnerability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZKDB enforces "Identity-before-Discovery," ensuring that even the metadata of a tool cannot be scraped for prompt injection.
* **Observability:** Tracks "Schema Unmasking" events in the Audit Log to detect abnormal mapping patterns.

## 7. Evolutionary Changelog
* **2026-07-01:** Initial Document Creation.
