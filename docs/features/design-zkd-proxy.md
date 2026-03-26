# Design Doc: Zero-Knowledge Discovery (ZKD) Proxy
**Status:** Draft
**Created:** 2026-06-27

## 1. Context and Scope
In modern agent swarms, the tool discovery phase has become a significant liability. Traditionally, agents broadcast their entire tool schema (including parameter descriptions and examples) to the registry, which then serves it to any discovering peer. This "Open Schema" model is vulnerable to **Pre-Flight Shadow Mapping**, where a malicious subagent scans the discovery bus to identify high-trust tools and subsequently injects "Shadow Capabilities" with identical names to intercept calls.

The Zero-Knowledge Discovery (ZKD) Proxy evolves the discovery process by mandating **Zero-Knowledge Capability Proofs (ZKCP)**. Agents prove they possess a specific skill (e.g., "Filesystem Writer") via a cryptographic commitment without revealing the underlying JSON-RPC schema or parameter definitions until a mission-bound, hardware-attested handshake is completed.

## 2. Goals & Non-Goals
* **Goals:**
    * Mask sensitive tool schemas from unauthenticated peers during the discovery phase.
    * Prevent "Shadow Mapping" by making capabilities cryptographically invisible without a valid mission token.
    * Facilitate "Capability Verification" where a requester can confirm a peer's skill level without seeing the implementation details.
* **Non-Goals:**
    * Replacing the A2A handshake (ZKD is the pre-flight stage for the handshake).
    * Restricting tool discovery for legitimate, authenticated mission roots.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Discover a "Database Administrator" subagent within a heterogeneous mesh without exposing the database schema to the entire discovery bus.
* **The Happy Path (Tasks):**
    1. A "Lead Agent" queries the ZKD Proxy for agents with `capability:db_admin`.
    2. The ZKD Proxy returns a list of "Encrypted Agent Cards" containing ZK-Proofs of skill possession.
    3. The Lead Agent selects a candidate and initiates an A2A Handshake using its hardware-bound Mission Root token.
    4. Upon successful attestation, the ZKD Proxy releases the full `db_admin` schema to the Lead Agent.
    5. The Lead Agent delegates the task, confident that the schema was never leaked to unauthorized "Shadow" agents on the bus.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>ZKD Proxy: Discovery Request (Skill: Git)
        ZKD Proxy->>Registry: Fetch Commitments
        Registry-->>ZKD Proxy: ZK-Proofs (Skill: Git)
        ZKD Proxy-->>Agent: Proved Capabilities (Masked Schemas)
        Agent->>ZKD Proxy: Handshake (Mission Token)
        ZKD Proxy->>ZKD Proxy: Validate Hardware Attestation
        ZKD Proxy-->>Agent: Full Tool Schema
    ```
* **APIs / Interfaces:**
    * `GET /v1/discovery/prove?skill=[name]`: Returns a list of agents providing ZK-Proofs for a skill.
    * `POST /v1/discovery/unmask`: Exchanges a mission-root token for a full capability schema.
* **Data Storage/State:**
    * Capability Commitments are stored in the Namespace-Locked Registry.
    * Schema Blobs are held in an encrypted, origin-locked sidecar store.

## 5. Alternatives Considered
* **Plaintext Discovery with RBAC:** Rejected because RBAC still requires the schema to be "known" by the central registry, creating a single point of failure for metadata exfiltration.
* **Opaque UUIDs:** Rejected as it breaks the semantic reasoning of LLMs during the planning phase.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZKD ensures that even if the discovery bus is monitored, no imperative instructions (from tool descriptions) can be scraped for prompt injection.
* **Observability:** The "Discovery Audit Log" tracks proof-requests vs. schema-unmasks, highlighting potential mapping attempts.

## 7. Evolutionary Changelog
* **2026-06-27:** Initial Document Creation.
