# Design Doc: Mission-Root Attestation Receipt (MRAR) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms move from cloud-centralized environments to edge-resident meshes (e.g., local P2P agent networks in OpenClaw), the reliance on a continuous cloud connection for hardware attestation has become a critical performance and reliability bottleneck. Subagents frequently lose access to local tools when transient network issues prevent real-time verification of session tokens against the cloud mission root.

The **Mission-Root Attestation Receipt (MRAR) Provider** solves this by introducing a "Receipt-of-Trust" model. It allows the MCP Any gateway to issue lightweight, hardware-signed, and time-bound receipts that subagents can present to local tools as proof of mission-bound authority. These receipts are verifiable offline, ensuring mission continuity in restricted network environments.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, mission-bound trust receipts to subagents.
    * Enable offline verification of subagent authority by local tool adapters.
    * Enforce monotonic receipt sequencing to prevent replay attacks.
    * Support mission-root bound capability scoping within the receipt.
* **Non-Goals:**
    * Replacing real-time cloud attestation for high-stakes remote operations.
    * Managing the full lifecycle of agent frameworks (OpenClaw, Claude, etc.).
    * Providing long-term storage for mission reasoning traces.

## 3. Critical User Journey (CUJ)
* **User Persona:** Edge-Resident Specialist Agent
* **Primary Goal:** Prove mission-bound authority to a local `run_shell_command` tool while the host is offline.
* **The Happy Path (Tasks):**
    1. The parent agent initiates a mission and requests an MRAR from MCP Any.
    2. MCP Any validates the request against the TPM and generates a signed receipt.
    3. The subagent receives the receipt and is delegated a task.
    4. The subagent presents the MRAR to the local tool adapter.
    5. The tool adapter verifies the TPM signature and monotonic counter offline.
    6. The tool executes the command within the scoped authority of the receipt.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] -> Request MRAR -> [MCP Any MRAR Provider] -> TPM Sign -> [MRAR]
    [MRAR] -> Present -> [Tool Adapter] -> Offline Verify (TPM PubKey + Monotonic) -> Execute
* **APIs / Interfaces:**
    * `POST /v1/mrar/mint`: Requests a new trust receipt (Input: mission_id, capabilities, duration).
    * `POST /v1/mrar/verify`: Local endpoint for tool adapters to verify receipts.
* **Data Storage/State:**
    * Active receipts are cached in the local encrypted SQLite state store.
    * TPM monotonic counters are tracked to prevent sequence hijacking.

## 5. Alternatives Considered
* **Persistent Trust Leases (LFTA):** Rejected because LFTA still requires periodic network heartbeats to verify the attestation revocation list (ARL). MRAR is designed for true offline-first continuity.
* **Shared Session Keys:** Rejected due to the risk of lateral movement if a single subagent process is compromised. MRARs are task-scoped and hardware-bound.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Receipts are cryptographically bound to the hardware (TPM) and the specific mission-root. Replay is prevented via monotonic counters.
* **Observability:** Receipt issuance and verification events are logged to the local security audit hub with hardware-bound timestamps.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
