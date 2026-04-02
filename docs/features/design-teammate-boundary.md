# Design Doc: Teammate Boundary Enforcer (TBE)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Shadow AI" lures targeting horizontal agent swarms (Claude Code Agent Teams) has exposed a critical vulnerability in mesh-based coordination. Specialist agents often share a common execution environment or broad mission-root tokens, allowing a single compromised teammate to achieve lateral movement and access unauthorized tools or sensitive context shards. MCP Any needs a strict **Teammate Boundary Enforcer (TBE)** to mandate Zero-Trust segmentation within the swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement per-teammate capability white-listing.
    * Enforce strict isolation between specialist context shards.
    * Mandate hardware-attested identity handshakes for all inter-teammate coordination.
    * Automatically revoke teammate capabilities if reasoning drift is detected by the AIR Hub.
* **Non-Goals:**
    * Managing top-level user authentication (handled by LOWA).
    * Enforcing global mission budgets (handled by RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars to the "Researcher" specialist.
* **The Happy Path (Tasks):**
    1. The Orchestrator defines a mission with 3 specialists: Coder, Researcher, and Executor.
    2. MCP Any issues session-bound, hardware-attested identity tokens for each.
    3. The "Researcher" agent attempts to call a `fs:write` tool restricted to the "Coder".
    4. The **Teammate Boundary Enforcer** intercepts the call, verifies the "Researcher's" capability manifest, and blocks the request.
    5. The attempt is logged, and the "Researcher's" session is quarantined for review.

## 4. Design & Architecture
* **System Flow:**
    `[Teammate Request] -> [TBE Middleware] -> [Capability Manifest Check] -> [Identity Verification] -> [Tool/Shard Access]`
* **APIs / Interfaces:**
    * `POST /v1/mesh/register-teammate`: Registers a new specialist with a specific capability manifest.
    * `GET /v1/mesh/verify-access`: Internal check for tool/shard authorization.
* **Data Storage/State:**
    Teammate manifests are stored in the hardware-locked **Mission-Root Attestation Registry**.

## 5. Alternatives Considered
* **Global Least-Privilege:** Rejected because horizontal swarms require specialists to have different, overlapping capabilities that cannot be managed by a single static policy.
* **Container-per-Agent:** Effective but introduces prohibitive latency for high-frequency teammate coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** TBE is the primary mechanism for preventing lateral movement in compromised meshes.
* **Observability:** Visualized via the **Service Mesh Topology Monitor**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
