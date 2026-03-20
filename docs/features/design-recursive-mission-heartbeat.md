# Design Doc: Recursive Mission-Heartbeat Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
The emergence of "Recursive Mission Hijacking" (RMH) reveals a critical vulnerability in deep agent delegation chains. Malicious subagents can manipulate "Mission-Bound Heartbeats" to gradually exfiltrate parent mission constraints or bypass security policies while maintaining a valid, but shallow, cryptographic signature.

The **Recursive Mission-Heartbeat Provider** solves this by evolving the heartbeat from a point-to-point signal into a hardware-attested, lineage-bound integrity proof. This ensures that every step in a delegation chain (A->B->C) is held accountable to the original mission-root through a cryptographically linked chain of heartbeats.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographically signed heartbeats that include the full delegation lineage.
    * Mandate hardware-attested (TPM/SEP) validation for every level of the heartbeat chain.
    * Detect and terminate sub-missions that exhibit "Heartbeat Desync" or lineage spoofing.
    * Provide real-time telemetry of mission health and alignment across deep swarms.
* **Non-Goals:**
    * Replacing existing transport-layer security (mTLS).
    * Managing the semantic content of the mission (handled by AID Hub).
    * Long-term persistence of heartbeat history beyond the mission lifecycle.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Ensure a sub-mission 4 levels deep cannot deviate from the root mission intent without immediate detection.
* **The Happy Path (Tasks):**
    1. The root agent (A) spawns subagent (B) and issues an initial mission-root token.
    2. MCP Any generates a hardware-attested Root Heartbeat (RH-1).
    3. Agent B spawns Agent C, passing the RH-1 and creating a child heartbeat (CH-1) linked to RH-1.
    4. Subagent C issues a heartbeat. MCP Any validates that C's heartbeat contains a valid proof of B and A's attestation.
    5. If C attempts to "spoof" its parent B's state, the lineage-hash fails, and MCP Any terminates C's session.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant RootAgent
        participant SubAgent
        participant MCP_Any_Provider
        participant TPM

        RootAgent->>MCP_Any_Provider: Initialize Mission
        MCP_Any_Provider->>TPM: Sign Root Heartbeat (RH)
        TPM-->>MCP_Any_Provider: Signed RH Token

        SubAgent->>MCP_Any_Provider: Delegate to Child
        MCP_Any_Provider->>TPM: Hash RH + ChildID (CH)
        TPM-->>MCP_Any_Provider: Signed CH Token

        Note over MCP_Any_Provider: Continuous Validation: Verify CH -> RH Linkage
    ```
* **APIs / Interfaces:**
    * `/v1/mission/heartbeat/sign`: Accepts lineage tokens and returns a hardware-attested child heartbeat.
    * `/v1/mission/heartbeat/verify`: Validates a heartbeat token against its hardware-attested lineage.
* **Data Storage/State:**
    * Ephemeral "Mission-Lineage Store" (in-memory) to track active heartbeats and their relationships.

## 5. Alternatives Considered
* **Point-to-Point Heartbeats:** Rejected because they are susceptible to "Middleman Hijacking" where a compromised intermediate agent can mask a child's deviation.
* **Centralized Heartbeat Registry:** Rejected due to performance bottlenecks and the "Single Point of Failure" risk in decentralized swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage tokens are hardware-bound and session-specific. Replay is prevented via monotonic nonces included in the TPM signature.
* **Observability:** Heartbeat failures trigger P0 "Mission Integrity Alerts" in the UI Dashboard.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
