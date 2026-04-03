# Design Doc: Persistent Memory Mesh (PMM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents evolve from single-interaction tools to long-running mission specialists, maintaining state across process restarts, framework handoffs, and session interruptions has become a critical stability requirement. Current "Pure RAG" solutions often fail to preserve the reasoning nuances and "internal monologue" that anchor an agent to its mission root.

MCP Any needs to provide a standardized, hardware-attested continuity layer. The Persistent Memory Mesh (PMM) ensures that an agent's cognitive state is durable, encrypted, and cryptographically bound to the mission root, allowing for seamless resumption of complex tasks without "Context Amnesia."

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested (TPM/Secure Enclave) state snapshots for agents.
    * Ensure end-to-end encryption of cognitive state at rest.
    * Support cross-framework state handoffs (e.g., OpenClaw to AutoGen) via standardized BSH (Binary State Handoff) interfaces.
    * Automate mission resumption after system restarts or process failures.
* **Non-Goals:**
    * Replacing long-term vector memory (RAG). PMM is for session/mission continuity.
    * Storing large binary artifacts not related to agent state.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Orchestrator
* **Primary Goal:** Resume a 12-hour code migration mission across a system reboot without losing the agent's refined reasoning about dependency conflicts.
* **The Happy Path (Tasks):**
    1. The orchestrator enables PMM for a mission-root branch.
    2. MCP Any automatically captures periodic hardware-attested state snapshots during tool-call idle times.
    3. The host system undergoes a forced reboot.
    4. Upon restart, the orchestrator requests mission resumption.
    5. MCP Any verifies the hardware identity, decrypts the latest PMM snapshot, and restores the agent's context and blackboard state.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> tool_call -> MCP Any -> [PMM Snapshotter] -> Secure Enclave (TPM)`
    `Recovery -> MCP Any -> [PMM Loader] -> TPM Verification -> Agent Context Restored`
* **APIs / Interfaces:**
    * `POST /v1/pmm/snapshot`: Trigger manual state persistence.
    * `POST /v1/pmm/resume`: Recover state for a given mission_id.
    * `GET /v1/pmm/status`: Check health of persistence mesh.
* **Data Storage/State:**
    Encrypted BSH fragments stored in local SQLite-backed "Continuity Vault," keyed by hardware-derived session tokens.

## 5. Alternatives Considered
* **Pure Client-Side Persistence:** Rejected due to lack of hardware-attestation guarantees and difficulty in cross-framework synchronization.
* **Cloud-Only State Storage:** Rejected to maintain "Local Sovereignty" and reduce latency/privacy risks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** State fragments are encrypted with keys resident in Secure Enclaves. Unauthorized process access to snapshots results in immediate integrity failure.
* **Observability:** PMM status heartbeats and snapshot latency metrics are exported to the System Health Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
