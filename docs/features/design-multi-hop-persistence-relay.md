# Design Doc: Multi-Hop Persistence Relay
**Status:** Draft
**Created:** 2026-06-05

## 1. Context and Scope
As AI agent swarms grow in complexity, task delegation often spans multiple hops (Agent A -> Agent B -> Agent C). Currently, hardware-attested trust leases are frequently hop-limited, requiring redundant handshakes that account for up to 30% of reasoning latency. The Multi-Hop Persistence Relay (MHPR) allows hardware-attested trust to persist securely across deep delegation chains without strength degradation.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable hardware-attested trust leases to survive multi-hop delegations.
    * Reduce latency in deep swarms by eliminating redundant handshakes.
    * Maintain cryptographic proof of the entire delegation lineage.
* **Non-Goals:**
    * Eliminating hardware attestation entirely.
    * Managing the underlying transport layer (e.g., Named Pipes, WebSockets).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Architect
* **Primary Goal:** Execute a 5-hop agent delegation chain with sub-10ms per-hop security overhead.
* **The Happy Path (Tasks):**
    1. Root agent initiates a task and requests a Multi-Hop Trust Lease.
    2. Trust Broker issues a hardware-attested lease with a "Persistence Depth" metadata field.
    3. Root agent delegates to Subagent B, passing the lease.
    4. Subagent B verifies the lease locally using the MHPR logic without a new hardware signature.
    5. Subagent B delegates to Subagent C, maintaining the same lease.
    6. Subagent C executes the final tool call; the gateway validates the multi-hop lineage.

## 4. Design & Architecture
* **System Flow:**
  [TPM/SEP] -> [Trust Broker] -> (Multi-Hop Lease) -> [Agent A] -> [Agent B] -> [Agent C] -> [MHPR Validator] -> [Gateway]
* **APIs / Interfaces:**
    * `MHPR.Extend(lease, subagentID)`: Cryptographically appends a hop to the lease.
    * `MHPR.Verify(lease)`: Validates the entire chain of custody.
* **Data Storage/State:**
    * Short-lived cache of active Multi-Hop Leases to speed up validation.

## 5. Alternatives Considered
* **Per-Hop Re-Attestation:** Rejected due to prohibitive latency in deep swarms.
* **Global Trust Tokens:** Rejected as they lack the granular lineage required for Zero Trust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses aggregate signatures to ensure that any tampering in the delegation chain invalidates the lease.
* **Observability:** Tracks "Delegation Depth" and latency savings per swarm.

## 7. Evolutionary Changelog
* **2026-06-05:** Initial Document Creation.
