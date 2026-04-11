# Design Doc: Monotonic Mission Authorization (MMA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of heterogeneous swarms, a new exploit pattern has emerged: "Phantom Missions." Malicious subagents utilize stylometric mimicry to spoof mission-initiation signals, tricking the coordination bus into spawning new mission branches with elevated privileges. MCP Any needs a mechanism to ensure that every mission initiation is cryptographically bound to a sequential, hardware-attested lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a TPM-bound monotonic counter for all mission-initiation signals.
    * Ensure that "Phantom Missions" spawned via mimicry are detected and blocked.
    * Provide a non-repudiable audit trail for mission lineage.
* **Non-Goals:**
    * Implement linguistic analysis (handled by the SIV provider).
    * Enforce resource quotas (handled by the Reasoning-Budget Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent subagents from autonomously creating new, high-privilege mission roots.
* **The Happy Path (Tasks):**
    1. The parent agent sends a `spawn_sub_mission` request to MCP Any.
    2. MCP Any requests a monotonic signature from the hardware enclave (TPM/SEP).
    3. The TPM increments the mission-root counter and signs the initiation token.
    4. MCP Any validates the signature and counter sequence before establishing the new mission branch.
    5. Subagents attempting to mimic this signal without access to the hardware counter are interdicted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MMA Broker: Request Mission Spawn
        MMA Broker->>HardwareEnclave: Get Monotonic Counter + Sign
        HardwareEnclave-->>MMA Broker: Signed MMA Token (Counter: N)
        MMA Broker->>MissionRegistry: Register Branch (Link: N-1 -> N)
        MissionRegistry-->>Agent: Sub-Mission Established
    ```
* **APIs / Interfaces:**
    * `POST /v1/mission/authorize`: Generates a hardware-attested initiation token.
    * `X-MCP-MMA-Token`: Header for all inter-agent mission initiation requests.
* **Data Storage/State:**
    * The monotonic counter is persisted in secure, non-volatile memory within the hardware enclave.

## 5. Alternatives Considered
* **Time-based tokens:** Rejected due to clock-skew vulnerabilities and "Replay-as-Initiation" attacks.
* **Pure Stylometric Verification:** Rejected as it can be bypassed by RL-optimized mimicry.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The MMA token is the root of trust for all subsequent sub-delegations.
* **Observability:** Every MMA event is logged with its unique monotonic ID, allowing for perfect lineage reconstruction.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
