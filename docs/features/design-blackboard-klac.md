# Design Doc: Key-Level Access Control (KLAC) for Blackboard
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
The Shared KV Store (Blackboard) is the primary mechanism for state synchronization in multi-agent swarms. However, the discovery of CVE-2026-35102 revealed that subagents can "shadow" or override global blackboard keys by injecting mission-local variants, leading to instruction poisoning.

KLAC evolves the Blackboard from a simple flat storage to a mission-bound, cryptographically owner-aware state engine. It ensures that every key-value pair is tied to a specific mission root and can only be mutated by authorized descendants.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind Blackboard key ownership to the Mission-Root Identity (TPM-signed).
    * Enforce "Write-Once-by-Owner" or "Authorized-Mutation" policies at the key level.
    * Prevent mission-local subagents from overriding parent-defined keys without explicit delegation.
* **Non-Goals:**
    * Implementing the underlying storage engine (handled by SQLite/memfd).
    * Managing inter-agent transport security (handled by LOWA/mTLS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a malicious specialist agent from overriding the `system.proxy.config` key in the shared blackboard.
* **The Happy Path (Tasks):**
    1. Parent agent writes `system.proxy.config` to the Blackboard using its mission-root token.
    2. KLAC records the Parent's hardware-attested identity as the owner of this key.
    3. Specialist subagent attempts to overwrite `system.proxy.config` with its own mission-local token.
    4. KLAC detects the identity mismatch and blocks the write, returning a `403 Forbidden: Ownership Violation` error.
    5. The Parent agent is notified of the shadowing attempt via the `MeshSecurityMonitor`.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Request] -> [Blackboard Middleware] -> [KLAC Validator] -> [ARI Provider (Ownership Check)] -> [Blackboard Engine]`
* **APIs / Interfaces:**
    * `PUT /blackboard/v2/keys/{key}`: Enhanced endpoint requiring an `X-Mission-Token`.
    * `GET /blackboard/v2/keys/{key}/metadata`: Returns ownership and lineage info for a key.
* **Data Storage/State:**
    * Blackboard schema is extended with `owner_lineage_id` and `signature_hash` columns for every row.

## 5. Alternatives Considered
* **Namespace Isolation (Rejected):** Providing every agent with a separate namespace. Rejected because swarms require "Shared State" to coordinate; namespaces prevent legitimate collaborative updates.
* **Global Immutable Keys (Rejected):** Making all parent keys immutable. Rejected because it prevents legitimate task-handoffs where a subagent *must* update a parent key (e.g., status flags).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Key ownership is non-transferable without a parent-signed delegation token.
* **Observability:** "Key Ownership Heatmap" in the UI to visualize which agent owns which parts of the mission state.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
