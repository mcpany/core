# Design Doc: Mesh-Resident Cognitive Load Balancing (MCLB)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Deep agent swarms often suffer from "Reasoning Stall" where a single parent agent becomes a bottleneck for coordinating sub-tasks. MCLB allows MCP Any to dynamically redistribute cognitive tasks (reasoning-intensive tool calls) across the agent mesh based on real-time capacity.

## 2. Goals & Non-Goals
* **Goals:**
    * Collect real-time latency and "Cognitive Load" (tokens/sec) metrics from connected agents.
    * Provide a redistribution broker for `TeammateTool` requests.
* **Non-Goals:**
    * Automatically spawning new agent instances (handled by the Orchestrator).
    * Modifying agent internal reasoning paths.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Scale Agentic Developer
* **Primary Goal:** Maintain low-latency swarms even when 50+ specialized subagents are active.
* **The Happy Path (Tasks):**
    1. Subagents report "Heartbeat" metrics to MCLB.
    2. Orchestrator submits a high-volume task list.
    3. MCLB routes tasks to the "coolest" (lowest load) authorized subagents.

## 4. Design & Architecture
* **System Flow:**
    `[Agents] --Metrics--> [MCLB Hub] <--Task Request-- [Orchestrator]`
* **APIs / Interfaces:**
    * `GET /mclb/mesh-load`: Retrieve the current load map.
    * `POST /mclb/route`: Negotiate the best teammate for a task.
* **Data Storage/State:**
    In-memory Load Map with 10-second TTL for freshness.

## 5. Alternatives Considered
* **Round-Robin Routing:** Rejected because it ignores the actual reasoning complexity of the tasks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Routing is only permitted between agents with a shared RIA mission token.
* **Observability:** Mesh load dashboards are integrated into the UI.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
