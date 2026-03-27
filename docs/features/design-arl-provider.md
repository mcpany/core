# Design Doc: ARL (Attestation Revocation List) Provider
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
With the introduction of "Trust Leases" (LFTA) in the agent ecosystem, agents can now perform tool calls without per-call hardware signatures. While this improves performance, it introduces a significant security window if an agent is compromised during the lease period. Traditional TTL-based expiration is too slow for machine-speed attacks.

The ARL Provider allows for sub-millisecond, hardware-bound revocation of agent capabilities across the entire MCP Any mesh. It acts as a real-time "Kill Switch" that broadcasts revocation signals to all connected nodes when a compromise is detected or reported by a trust-root.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate real-time broadcast of Attestation Revocation Lists (ARLs).
    * Implement sub-millisecond capability revocation at the transport layer.
    * Provide hardware-bound (TPM/SEP) verification for revocation signals.
    * Support federated revocation where multiple nodes can signal a compromise.
* **Non-Goals:**
    * Defining the criteria for compromise detection (handled by CSAD or external auditors).
    * Managing the initial issuance of trust leases (handled by the Trust Lease Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Operator
* **Primary Goal:** Instantly revoke all capabilities for an agent team identified as exhibiting "Hivenet" attack patterns.
* **The Happy Path (Tasks):**
    1. The CSAD Hub detects anomalous behavioral entropy in Swarm X.
    2. The operator (or an automated auditor) issues a "Revocation Request" to the ARL Provider.
    3. The ARL Provider signs the revocation signal with the mesh-root TPM key.
    4. The signal is broadcasted over the isolated inter-agent bus.
    5. MCP Any nodes immediately terminate all active sessions and leases associated with the revoked agent IDs.
    6. Any subsequent tool calls from the revoked agents are blocked at the kernel-pipe level.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Auditor[Auditor / CSAD Hub] -->|Compromise Detected| ARLProvider[ARL Provider]
        ARLProvider -->|Sign Revocation| TPM[Hardware TPM/SEP]
        TPM -->|Signed ARL| ARLProvider
        ARLProvider -->|Broadcast| MeshNodes[Mesh Nodes / Gateways]
        MeshNodes -->|Interdict| AgentTransport[Agent Transport / Pipes]
    ```
* **APIs / Interfaces:**
    * `POST /arl/revoke`: Issues a new revocation signal for an agent or session ID.
    * `GET /arl/list`: Returns the current active revocation list (signed).
    * `SUBSCRIBE /arl/events`: WebSocket/gRPC stream for real-time revocation updates.
* **Data Storage/State:**
    * The active ARL is held in a high-speed, replicated in-memory Bloom filter for sub-microsecond lookups during tool calls.

## 5. Alternatives Considered
* **Short-lived Leases (1s):** Rejected due to the "Attestation Tax" latency still impacting high-frequency reasoning.
* **Centralized Gatekeeping:** Rejected as it creates a single point of failure and bottleneck for distributed meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Revocation signals must be cryptographically bound to the hardware root to prevent "Denial of Agency" attacks by malicious subagents.
* **Observability:** Every revocation event is logged with high-fidelity stylometric context to the "Local Security Audit Log."

## 7. Evolutionary Changelog
* **2026-03-22:** Initial Document Creation.
