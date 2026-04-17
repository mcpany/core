# Design Doc: Recursive Swarm Heartbeats (RSH) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become deeper and more autonomous (e.g., OpenClaw's multi-level delegation), the mission root increasingly loses visibility into the "Cognitive Health" of sub-swarms. Traditional Liveness checks (ping/pong) only confirm process availability, not reasoning integrity. "Reasoning Loops" or "Cognitive Stalls" in specialist subagents can exhaust mission budgets without producing results.

The Recursive Swarm Heartbeat (RSH) Hub is required to provide a standardized, hardware-attested protocol for subagents to broadcast their real-time reasoning entropy, task progress, and intent-alignment back to the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a recursive heartbeat protocol for multi-level agent swarms.
    * Facilitate real-time reporting of reasoning entropy and semantic intent-alignment.
    * Enable the mission root to proactively trigger "Cognitive Load Shedding" or "Subagent Reaping" based on heartbeat health.
    * Ensure heartbeats are hardware-attested (TPM) to prevent "Ghost Reasoning" spoofing.
* **Non-Goals:**
    * Replacing the primary coordination mailbox (A2A/UACO).
    * Providing a general-purpose logging or telemetry sink.
    * Executing automated self-correction; RSH only provides the visibility for the mission root to act.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Governance Architect
* **Primary Goal:** Detect and terminate a specialist subagent that has entered an infinite refinement loop in a deep sub-mission.
* **The Happy Path (Tasks):**
    1. Mission root spawns a specialist subagent with an RSH policy.
    2. The subagent periodically generates an RSH packet containing its current reasoning entropy score and last 3 chain-of-thought summaries.
    3. The subagent signs the packet using its hardware-attested session token.
    4. The RSH Hub intercepts the packet and propagates it up the delegation chain.
    5. The RSH Hub analyzes the entropy trend; it detects a "Cognitive Stall" (flat entropy with high token consumption).
    6. The mission root receives a "Health Warning" and issues a termination signal to the stalled subagent via the Active Subagent Reaper.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent N] -->|Hardware-Signed RSH| B(RSH Hub)
        B -->|Recursive Propagation| C[Parent Agent]
        C -->|Aggregated Health| D[Mission Root]
        B -->|Entropy Analysis| E{Threshold Check}
        E -->|Stall Detected| F[Load Shedding Trigger]
    ```
* **APIs / Interfaces:**
    * `rsh.BroadcastHeartbeat(agentID, healthMetrics, CoTSummaries) -> Ack`: Subagent reports health.
    * `rsh.GetSwarmHealth(missionRootID) -> HealthMap`: Mission root queries recursive health status.
    * `rsh.SetThreshold(policyID, entropyLimit) -> Status`: Configures stall detection parameters.
* **Data Storage/State:**
    * **Swarm Health Registry:** In-memory store of active agent heartbeats and their last-seen health states.
    * **Entropy Analytics Buffer:** Short-term window of metrics for trend analysis.

## 5. Alternatives Considered
* **Standard OpenTelemetry Sinks:** Rejected because they lack "Agentic Context" and cannot enforce hardware-bound attestation for the reasoning traces themselves.
* **Synchronous Health Checking (Polling):** Rejected as it increases coordination latency and doesn't scale for large horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All heartbeats must be cryptographically linked to the agent's mission-bound identity. Spoofed heartbeats trigger an immediate mission-wide quarantine.
* **Observability:** Integrated with the "Agentic Entropy Scoreboard" in the UI for real-time visualization of swarm health.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
