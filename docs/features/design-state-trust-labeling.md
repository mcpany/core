# Design Doc: State-Trust Labeling (STL) Provider
**Status:** Draft
**Created:** 2026-05-19

## 1. Context and Scope
As AI agent swarms become more heterogeneous, bridging between disparate frameworks (UAB, A2A, MCP), they become vulnerable to **Protocol-Agnostic State Injection (PASI)**. This occurs when an agent ingests state from a lower-trust source (e.g., an unauthenticated legacy MCP server) and unknowingly propagates it into a high-trust reasoning session or the Shared KV Store (Blackboard). MCP Any needs to solve this by implementing mandatory "State-Trust Labeling" for all data fragments.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a labeling mechanism for all KV data in the Blackboard, tagging it with the trust level of its origin.
    * Enforce trust-based read/write policies (e.g., high-trust reasoning loops cannot ingest low-trust data without explicit "Sanitization").
    * Integrate with hardware-attested identity tokens to provide verifiable trust origins.
* **Non-Goals:**
    * Performing semantic validation of the data itself (handled by other sanitizers).
    * Restricting all data sharing (focus is on labeling and policy enforcement).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent an unauthenticated subagent from poisoning the "Mission State" shared by a verified teammate.
* **The Happy Path (Tasks):**
    1. The orchestrator defines trust-level policies in the MCP Any configuration.
    2. A verified Claude Code teammate writes a "Mission Success" flag to the Blackboard; it is tagged as `TRUST_LEVEL_HIGH`.
    3. An unauthenticated legacy MCP tool attempts to overwrite the same key; the STL Provider intercepts and rejects the write due to a `TRUST_LEVEL_LOW` origin.
    4. A monitor agent attempts to read the state; the STL Provider ensures it only receives data fragments that meet its required trust threshold.

## 4. Design & Architecture
* **System Flow:**
    `[Data Source] -> [STL Provider] -> [Blackboard (SQLite)]`
* **APIs / Interfaces:**
    * `STL.tag_data(key, value, origin_token)`: Computes and attaches a trust label based on the origin's attestation strength.
    * `STL.verify_access(requester_token, target_label)`: Evaluates if the requester has sufficient trust to ingest the data.
* **Data Storage/State:**
    * Trust labels are stored as metadata in a sidecar table for the Blackboard SQLite database.

## 5. Alternatives Considered
* **Namespace Isolation**: Rejected as it prevents legitimate cross-agent state sharing which is required for swarm coordination.
* **Encryption-at-Rest**: Complements STL but doesn't solve the "Trust Hijacking" problem where low-trust agents can still write valid, but malicious, plaintext.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: STL is a core pillar of the Zero Trust architecture, ensuring data lineage is preserved.
* **Observability**: Trust-policy violations will be surfaced in the Local Security Audit Dashboard.

## 7. Evolutionary Changelog
* **2026-05-19:** Initial Document Creation.
