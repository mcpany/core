# Design Doc: Reason-Graph Integrity (RGI) Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Autonomous agent swarms are increasingly vulnerable to "Reason-Graph Collision" (RGC) exploits. In these attacks, malicious subagents inject circular or conflicting reasoning nodes into a shared mesh, triggering cognitive deadlocks and intent eviction. The "Universal Agent Bus" must evolve beyond simple context isolation to protect the structural integrity of the reasoning path itself.

The Reason-Graph Integrity (RGI) Provider acts as the authoritative validator for all inter-agent reasoning traces, ensuring that the swarm's cognitive path remains acyclic and mission-anchored.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested structural analysis for reasoning graphs.
    * Detect and block circular reasoning (RGC) in sub-millisecond real-time.
    * Mandate hardware-locked lineage for every reasoning node in the graph.
* **Non-Goals:**
    * Validating the truthfulness of the reasoning (handled by VRP).
    * Enforcing network-layer encryption (handled by T2T Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent an OpenClaw specialist from stalling the mission-root via circular reasoning injection.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a task to an OpenClaw specialist.
    2. Specialist proposes a reasoning trace back to the mesh.
    3. RGI Provider intercepts the trace and performs structural cycle detection.
    4. RGI Provider verifies the hardware-locked lineage of each proposed node.
    5. Trace is approved and committed to the shared reason-graph.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Proposal] --> B[RGI Middleware]
        B --> C[Cycle Detection Engine]
        C --> D[Lineage Authenticator]
        D --> E{Integrity Check}
        E -- Pass --> F[Reason-Graph Commit]
        E -- Fail --> G[Isolate Subagent]
    ```
* **APIs / Interfaces:**
    * `rgi.ValidateTrace(trace_id, nodes[])`: Analyzes structural integrity of a reasoning fragment.
    * `rgi.CheckLineage(node_id)`: Verifies hardware attestation of a reasoning node.
* **Data Storage/State:**
    * Reasoning traces are stored in a hardware-isolated, acyclic state-graph on the Blackboard.

## 5. Alternatives Considered
* **Time-based Deadlock Detection:** Rejected because RGC can mimic active reasoning while being semantically stagnant.
* **Manual Review:** Rejected due to machine-speed coordination requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RGI is anchored to TPM-bound session identities.
* **Observability:** Graph violations are logged as P0 security events in the Audit Log.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
