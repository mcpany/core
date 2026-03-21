# Design Doc: Mesh-Resident Cognitive Load Balancer (MCLB)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Deep agent swarms often encounter "Cognitive Stall" or "Reasoning Bottlenecks" where high-intensity specialist agents are over-utilized while others remain idle. This leads to increased mission latency and resource inefficiency.

The MCLB is a mesh-resident service that monitors the real-time reasoning load (ARE headers, context window utilization) of all active agents in the mesh. It dynamically redistributes reasoning tasks to available teammates with compatible capability cards, ensuring optimal throughput and preventing hangups in deep multi-hop chains.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time monitoring of agent reasoning effort.
    * Provide dynamic redistribution of task assignments based on cognitive capacity.
    * Minimize "Negotiation Exhaustion" by automating load-based bidding.
* **Non-Goals:**
    * It will NOT perform task execution (handled by specialized agents).
    * It will NOT enforce security policies (handled by RBF/ISD).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cloud-Native Swarm Architect
* **Primary Goal:** Maintain low-latency performance for a high-concurrency 50-agent swarm during a burst of complex reasoning requests.
* **The Happy Path (Tasks):**
    1. Multiple specialist agents receive high-entropy reasoning fragments.
    2. The MCLB detects that 3 agents have reached 90% context utilization.
    3. The MCLB identifies 5 idle agents with compatible "Skill Cards."
    4. The MCLB triggers a "Cognitive Handoff," migrating sub-tasks to the idle agents.
    5. Reasoning continues in parallel, neutralizing the bottleneck.

## 4. Design & Architecture
* **System Flow:**
    [Agent Mesh] -> [ARE Telemetry] -> [MCLB Service] -> [Task Migration Signal] -> [Agent Handoff]
    The MCLB operates as a high-speed telemetry sink and coordination arbiter.
* **APIs / Interfaces:**
    * `ReportLoad(agentID, loadMetrics)`: Ingests real-time performance data.
    * `BalanceTask(taskCard) -> targetAgentID`: Determines the optimal recipient for a delegated task.
* **Data Storage/State:**
    Uses a lock-free, mesh-resident state shard for sub-millisecond coordination.

## 5. Alternatives Considered
* **Static Round-Robin:** Rejected as it doesn't account for the non-deterministic reasoning effort of LLMs.
* **Agent-Managed Balancing:** Rejected due to the coordination overhead and risk of "Negotiation Storms."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All task migrations must pass SMI/RIA attestation to ensure the target agent is authorized for the mission context.
* **Observability:** MCLB performance (balancing events, latency reduction) is visualized in the UI Lineage Explorer.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
