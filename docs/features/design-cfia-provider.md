# Design Doc: Continuous Fragment-Integrity Attestation (CFIA)
**Status:** Draft
**Created:** 2026-06-17

## 1. Context and Scope
With the shift toward horizontal teammate coordination in heterogeneous swarms (e.g., Claude Code Agent Teams interacting with OpenClaw specialists), shared state shards have become the primary attack surface. The emergence of **Mesh-Resident Logic Bombs (MRLB)**—dormant malicious state fragments that trigger only upon specific mission-root state shifts—requires a transition from point-in-time validation to continuous, lifecycle-wide attestation.

The **Continuous Fragment-Integrity Attestation (CFIA)** service provides a cryptographically signed proof of shard-level integrity that is re-attested periodically throughout the mission lifecycle.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Perform periodic, hardware-attested re-validation of shared state shards.
    *   Detect and neutralize dormant "Logic Bombs" before they can be triggered.
    *   Provide fragment-level integrity proofs for all A2A-compliant teammates.
    *   Support sub-millisecond re-attestation to avoid mesh latency.
*   **Non-Goals:**
    *   Replacing the initial shard-level validation (ARI).
    *   Modifying the underlying LLM's reasoning engine.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Maintain mission-root sovereignty over a shared teammate shard for a long-running, multi-agent task.
*   **The Happy Path (Tasks):**
    1.  The agent initializes a shared teammate shard with an ARI signature.
    2.  The CFIA service is activated for the mission scope.
    3.  CFIA performs periodic, hardware-bound (TPM) scans of the shard.
    4.  If a "Logic Bomb" pattern is detected (semantic drift from mission-root baseline), CFIA revokes the shard's integrity token.
    5.  Teammates are automatically blocked from ingesting the compromised fragment.
    6.  Mission-root remains sovereign despite the presence of dormant malicious state.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        ARI[ARI Validator] -->|Initial Sign-off| Shard[Shared State Shard]
        CFIA[CFIA Provider] -->|Periodic Scan| Shard
        CFIA -->|Semantic Baseline| MissionRoot[Mission Root Intent]
        CFIA -->|Integrity Proof| Teammates[Teammate Agents]
        Teammates -->|Request Fragment| Shard
        Shard -->|Token Check| CFIA
    ```
*   **APIs / Interfaces:**
    *   `AttestShard(shardID string, missionRoot Identity) (Proof, error)`: Core internal function for periodic fragment validation.
    *   `x-mcpany-cfia-token`: New header for transporting real-time integrity proofs in inter-agent messages.
*   **Data Storage/State:**
    *   Utilizes a local, hardware-locked manifest of all active state shards and their initial ARI baselines.

## 5. Alternatives Considered
*   **One-Time Validation (ARI Only):** Rejected because it cannot detect "Logic Bombs" that only become malicious as the mission-root state evolves over time.
*   **Global Mailbox Locks:** Rejected due to the performance tax on parallel teammate coordination.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** CFIA requires hardware-bound (TPM) signatures to prevent spoofing of integrity proofs.
*   **Observability:** Integrated with the `Swarm Anomaly Visualizer`, showing real-time "Integrity Trends."

## 7. Evolutionary Changelog
*   **2026-06-17:** Initial Document Creation.
