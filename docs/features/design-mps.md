# Design Doc: Mesh Policy Synchronizer (MPS)
**Status:** Draft | In Review | Approved
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of large-scale agent swarms (e.g., CrewAI, AutoGen), ensuring that 100+ parallel agents are operating under the same set of security guardrails is a major challenge. The MPS provides a "Single Source of Truth" for security policies across a distributed agent mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Sub-10ms "Policy Heartbeat" to ensure all agents have the latest security constraints.
    * Atomic policy updates across heterogeneous agent frameworks (OpenClaw, CrewAI, etc.).
    * Hardware-attested policy versioning.
* **Non-Goals:**
    * Real-time enforcement of the policies (this is the job of the individual agent adapters).
    * Centralized storage of long-term agent history (MPS focuses on current policy state).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer for a Corporate Agent Swarm
* **Primary Goal:** Revoke "File Write" permissions from all 50 active agents within 10ms of a detected breach.
* **The Happy Path (Tasks):**
    1. Security Engineer updates a Rego-based policy via the MCP Any UI.
    2. MPS detects the policy change.
    3. MPS broadcasts a "Policy Heartbeat" to all registered agents.
    4. Each agent's adapter (via MCP Any) receives the new policy fragment.
    5. Agent #34 attempts a "File Write" but is blocked by the new local policy state.

## 4. Design & Architecture
* **System Flow:**
    `Security UI -> MPS (Central Registry) -> [Policy Heartbeat] -> [Agent 1, Agent 2, ... Agent N]`
* **APIs / Interfaces:**
    `/v1/mesh/policy_sync`, `/v1/mesh/heartbeat`
* **Data Storage/State:**
    Active policies are stored in a high-speed, in-memory KV store (e.g., Redis) with hardware-attested versioning.

## 5. Alternatives Considered
* **Periodic Polling:** Rejected because the latency (typically >1s) is too slow for security-critical revocations.
* **Client-Pull:** Rejected due to "Policy Drift" where individual agents might not pull updates during high CPU load.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All "Policy Heartbeats" are signed with the MCP Any root certificate to prevent "unauthorized policy injection."
* **Observability:** MPS sync status is tracked in the `/ui/mesh-policy-editor`.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
