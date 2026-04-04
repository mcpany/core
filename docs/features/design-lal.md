# Design Doc: Lightweight Attestation Lease (LAL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent meshes grow in density and horizontal scaling (e.g., Claude Code Agent Teams), the "Handshake Exhaustion" phenomenon has emerged as a primary performance bottleneck. Standard TPM-bound hardware handshakes, while secure, introduce a 100ms+ latency penalty per cross-node tool call or state synchronization event. In a mesh with 10+ subagents, this coordination tax exceeds the actual tool execution time. LAL provides a mechanism to issue time-bound, cryptographically derived "Lightweight Identities" that maintain hardware-rooted trust while enabling sub-10ms inter-agent coordination.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-node coordination latency from 100ms+ to <10ms.
    * Provide hardware-rooted (TPM/SEP) lineage for all lightweight leases.
    * Enable automatic, non-blocking lease rotation across deep meshes.
    * Neutralize "Handshake Exhaustion" in high-density horizontal swarms.
* **Non-Goals:**
    * Replacing full hardware attestation for initial mission-root establishment.
    * Providing long-term persistent identity (LALs are ephemeral by design).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate a 12-agent "Refactoring Swarm" across local Docker nodes without hitting coordination stalls.
* **The Happy Path (Tasks):**
    1. Parent agent establishes a hardware-attested session with the MCP Any gateway.
    2. Parent agent requests a "Lease Package" for its 12 subagents.
    3. LAL Provider issues a TPM-signed Root Lease and 12 derived Sub-Leases.
    4. Subagents use LAL tokens for sub-millisecond mTLS handshakes with sibling nodes.
    5. MCP Any monitors lease expiration and speculatively prepares a "Rotation Package" to prevent cognitive stall.

## 4. Design & Architecture
* **System Flow:**
    `Hardware Root (TPM)` -> `LAL Master Key` -> `Ephemeral Session Keys` -> `Subagent Tokens (JWT-BSH)`
* **APIs / Interfaces:**
    * `LeaseService`: `MintLease(missionID string, duration time.Duration) (LeaseBundle, error)`
    * `LeaseValidator`: `VerifyDerivedToken(token string) (Claims, error)`
* **Data Storage/State:**
    * Lease metadata is stored in the `Asynchronous Mailbox Shards` (AMS) to ensure lock-free access during rotation.

## 5. Alternatives Considered
* **Persistent mTLS Certificates**: Rejected due to high management overhead and risk of long-term key compromise in specialist subagent processes.
* **Centralized Auth Proxy**: Rejected because it introduces a single point of failure and additional network hops that negate the latency benefits.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** LALs are mathematically bound to the hardware-attested mission root. Any attempt to use a lease outside its mission-scope triggers a hardware-level invalidation signal.
* **Observability:** Lease minting and rotation events are tracked in the "Mesh Identity Manager" UI component.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
