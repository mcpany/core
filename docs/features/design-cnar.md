# Design Doc: Cross-Node Attention Resumption (CNAR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agentic missions scale to millions of tokens and migrate between local execution environments and cloud-based reasoning engines, "Context Amnesia" has become a primary bottleneck. Agents lose their attention state when switching physical nodes, leading to expensive re-computation of KV caches and loss of subtle reasoning nuances.

CNAR provides a hardware-attested mechanism to serialize, migrate, and resume "Attention Snapshots" across disparate physical nodes, ensuring cognitive continuity for long-haul missions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Standardize the format for hardware-attested "Attention Snapshots."
    *   Provide a secure coordination layer for migrating snapshots between physical MCP Any nodes.
    *   Integrate with TPM/Secure Enclave to sign the "Lineage of Attention."
*   **Non-Goals:**
    *   Implementing model-specific KV-cache optimization (handled by inference engines).
    *   Providing a global persistent storage for all agent state (focused on Attention only).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Hybrid-Cloud Agent Architect
*   **Primary Goal:** Migrate a 2.5M token reasoning session from a local developer machine to a high-memory cloud cluster for final synthesis without losing the "Chain of Thought."
*   **The Happy Path (Tasks):**
    1.  Agent completes a local reasoning phase and triggers a `CNAR_SNAPSHOT` event.
    2.  Local MCP Any instance generates a TPM-signed manifest of the current attention window.
    3.  The snapshot is encrypted and transmitted to the Target Node via an **Attested Mesh Tunnel (AMT)**.
    4.  Target Node verifies the hardware signature and the mission-root lineage.
    5.  Cloud-based LLM resumes reasoning using the pre-flight attention map, providing sub-second continuity.

## 4. Design & Architecture
*   **System Flow:**
    [Local Node] --(TPM Sign)--> [Attention Snapshot]
                                         |
    [Attested Mesh Tunnel] <-------------'
          |
    [Target Node] --(Verify)--> [Inference Engine]

*   **APIs / Interfaces:**
    *   `POST /v1/continuity/snapshot/create`: Generates a signed attention map.
    *   `POST /v1/continuity/snapshot/resume`: Ingests and verifies a remote snapshot.
*   **Data Storage/State:**
    *   Snapshots are transient and held in encrypted `memfd` segments during migration.

## 5. Alternatives Considered
*   **Context Re-ingestion:** Rejected due to the massive token cost and latency of re-processing 2M+ tokens.
*   **Centralized State Hub:** Rejected due to data privacy concerns for sensitive local context.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Snapshots are cryptographically bound to the Mission Root. A stolen snapshot cannot be resumed without the corresponding hardware-attested session key.
*   **Observability:** Tracked via the **Attention Snapshot Manager** (UI Roadmap).

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
