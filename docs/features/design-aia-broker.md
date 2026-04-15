# Design Doc: Active Intent Alignment (AIA) Broker
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
As AI agent swarms grow in depth and complexity, "Intent Drift" has emerged as a primary failure mode. specialist agents, while following their specific sub-tasks, can gradually deviate from the parent agent's mission root, leading to "Cognitive Hallucinations" or unauthorized tool execution. Current state-sharing mechanisms (like the Blackboard) are passive and do not enforce semantic consistency over time.

The Active Intent Alignment (AIA) Broker is required to provide an authoritative alignment service that issues hardware-attested "Alignment Heartbeats." It ensures that specialist agent reasoning traces remain semantically anchored to the mission root, neutralizing cumulative drift in deep horizontal meshes.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a background monitoring service for subagent reasoning monologues.
    * Provide hardware-attested "Alignment Heartbeats" (TPM-signed semantic checks).
    * Automatically trigger "Correction Cycles" or mission termination upon detection of significant intent drift.
    * Neutralize "Mirror-Splice" attacks by verifying alignment before re-injecting recovery mirrors.
* **Non-Goals:**
    * Modifying the subagent's internal reasoning logic.
    * Restricting subagent autonomy within its authorized mission scope.
    * Managing the transport layer for agent communications.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Swarm Architect
* **Primary Goal:** Prevent a specialist "Database Refactoring" agent from "drifting" into unauthorized "Schema Deletion" during a complex multi-hour migration.
* **The Happy Path (Tasks):**
    1. A complex migration mission is initiated with a hardware-attested mission root.
    2. AIA Broker starts monitoring the specialist agent's reasoning traces via the Blackboard.
    3. Every 30 seconds, the Broker performs a semantic comparison between the current reasoning and the mission root.
    4. If the Broker detects a "Semantic Divergence" (e.g., the agent starts speculating about table deletion), it issues a `DRIFT_DETECTED` signal.
    5. The mission is automatically paused, and the "Self-Correction Loop Arbiter" is triggered to realign the agent.
    6. Once realigned, the AIA Broker issues a new TPM-signed "Alignment Heartbeat," and the mission resumes.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[AIA Broker]
        C[Subagent Monologue] --> D[Blackboard]
        D --> B
        B -->|Semantic Analysis| E{Aligned?}
        E -->|Yes| F[Issue Heartbeat]
        E -->|No| G[Trigger Correction]
        F --> H[Continue Mission]
    ```
* **APIs / Interfaces:**
    * `aia.MonitorIntent(missionID, subagentID)`: Registers a subagent for alignment monitoring.
    * `aia.CheckAlignment(monologue, rootIntent) -> Score`: Performs semantic comparison.
    * `aia.GetHeartbeat(missionID) -> HeartbeatToken`: Returns the latest hardware-attested proof of alignment.
* **Data Storage/State:**
    * **Alignment Registry:** In-memory map of active missions and their semantic baselines.
    * **Heartbeat Store:** Hardware-attested log of alignment checks.

## 5. Alternatives Considered
* **Periodic Manual Review:** Rejected as it is not scalable for autonomous swarms and leads to approval fatigue.
* **Hard-Coded Keyword Filtering:** Rejected because it fails to capture the nuanced intent of complex reasoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Alignment checks are performed in a privileged gateway layer. MSF (Mirror-Splice Firewall) integration ensures recovery buffers are also aligned.
* **Observability:** Integrated with the "Active Intent Alignment Monitor" in the UI for real-time visualization of drift scores and heartbeat status.

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
* **2026-07-25:** Mirror-Splice Integration.
    * Context: Mirror-Splice attacks attempt to bypass AIA by injecting drift into recovery buffers.
    * Architecture Adjustment: AIA Broker now performs a mandatory "Pre-Injection Alignment Check" for all Semantic Mirror recovery events.
    * Security Impact: Closes the "Recovery-Path Bypass" for intent drift.
