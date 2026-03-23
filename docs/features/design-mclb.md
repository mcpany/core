# Design Doc: MCLB (Mesh-Resident Cognitive Load Balancer)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In heterogeneous swarms, reasoning capacity varies wildly between agents (e.g., GPT-4o-based specialists vs. local-LLM monitors). Without coordination, "reasoning stalls" occur when high-intensity tasks are assigned to under-resourced agents. MCP Any needs to treat cognitive load as a dynamically balanced mesh resource.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor real-time cognitive utilization across the agent mesh.
    * Proactively redistribute task bidding (UACO) based on available ARE (Reasoning Effort) budgets.
    * Reduce overall swarm latency by minimizing "Refinement Drift" in low-capacity nodes.
* **Non-Goals:**
    * Automatically upgrading agent models (e.g., GPT-3 to GPT-4).
    * Managing compute/GPU allocation directly (handled at the model provider layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Maintain swarm throughput when a local specialist agent is overwhelmed by complex reasoning.
* **The Happy Path (Tasks):**
    1. MCLB detects a "Cognitive Stall" signal from a specialist agent (high latency, low reasoning confidence).
    2. MCLB queries the UACO registry for alternative agents with available ARE budget.
    3. MCLB triggers a "Task Re-Auction" for the stalled intent branch.
    4. A more capable (or less loaded) agent claims the task, restoring mesh performance.

## 4. Design & Architecture
* **System Flow:**
    `Agent Telemetry -> MCLB (Load Analysis) -> UACO Broker (Task Redistribution) -> Optimized Mesh Execution`
* **APIs / Interfaces:**
    * `GET /v1/mclb/metrics`: Retrieve mesh-wide cognitive utilization.
    * `POST /v1/mclb/rebalance`: Manually trigger a mesh rebalancing cycle.
* **Data Storage/State:** Real-time load metrics are stored in an in-memory, high-speed shard of the Blackboard.

## 5. Alternatives Considered
* **Static Priority Queuing**: Rejected because it cannot adapt to dynamic changes in reasoning complexity or model availability.
* **User-Managed Load Balancing**: Rejected due to the high MTTC (Mean Time To Coordinate) in autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Rebalancing triggers must be RIA-attested to prevent "Load-Shedding" DoS attacks.
* **Observability:** Load balancing dashboards are provided in the UI for real-time mesh monitoring.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
