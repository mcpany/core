# Design Doc: Multi-Hop Persistence Relay (MHPR)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
In complex, hierarchical agent swarms, a mission root often delegates tasks to subagents, which may further delegate to specialists (A -> B -> C). Current hardware-attested trust models often require a fresh handshake at every hop to maintain security posture. This "Handshake Tax" introduces significant latency (often 100ms+ per hop) and leads to "Cognitive Stall" in deep delegations.

The **Multi-Hop Persistence Relay (MHPR)** provides a mechanism for hardware-attested trust leases to persist across multiple delegation hops. It allows the security context of the mission root to be propagated through the swarm without redundant full handshakes, while maintaining cryptographic proof of lineage and sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable sub-10ms trust propagation across infinite delegation hops.
    * Facilitate "Trust Continuity" where subagents inherit the mission-root's attestation strength.
    * Mandate hardware-locked "Lineage Tokens" that are immutable during propagation.
    * Support real-time revocation of a relay chain if any hop is compromised.
* **Non-Goals:**
    * Eliminating the initial mission-root attestation.
    * Providing long-term archival of delegation traces (handled by SRM).
    * Managing tool-specific permissions (handled by DPG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Architect
* **Primary Goal:** Delegate a complex analysis task through 4 levels of specialized agents without the cumulative 400ms latency tax of redundant handshakes.
* **The Happy Path (Tasks):**
    1. The Mission Root (Agent A) performs a hardware-attested boot and establishes a "Persistent Trust Lease" with the MHPR.
    2. Agent A delegates to Agent B, attaching an MHPR "Relay Token" signed by the TPM.
    3. Agent B delegates to Agent C. MHPR validates the token against Agent A's root lease and extends it to Agent C in sub-millisecond time.
    4. Agent C executes a high-risk tool call.
    5. The Gateway verifies the MHPR chain back to Agent A's hardware-attested session and allows the call without requiring a new user signature.

## 4. Design & Architecture
* **System Flow:**
    `Mission Root (TPM) -> [Relay Token] -> Subagent -> [Chain Extension] -> Specialist -> [Chain Validation] -> Gateway`
* **APIs / Interfaces:**
    * `MHPR_Issue_Relay(lease_id, target_agent_id) -> relay_token`: Issue a trust relay token to a subagent.
    * `MHPR_Validate_Chain(chain_blob) -> root_lease_metadata`: Validate the complete lineage of a relay chain.
* **Data Storage/State:**
    * Chain states are held in memory-mapped shared regions for zero-latency lookup.
    * Authoritative lease registry backed by the `Sovereign Mesh Identity (SMI) Relay`.

## 5. Alternatives Considered
* **Recursive Handshaking**: Rejected due to prohibitive latency in deep swarms (O(n) latency tax).
* **JWT-based Bearer Tokens**: Rejected because they lack hardware-bound lineage and are susceptible to "Token Splicing" if the MCP Any process is compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MHPR uses "Hardware-Attested Intent Lineage" (HAIL) to ensure that relay tokens cannot be re-used outside their specific mission branch.
* **Observability:** Delegation latency and chain depth are visualized in the "Multi-Hop Trust Persistence Monitor."

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
