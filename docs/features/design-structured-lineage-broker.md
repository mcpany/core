# Design Doc: Structured Lineage Broker
**Status:** Draft
**Created:** 2026-03-27

## 1. Context and Scope
As agent swarms move from linear sessions to complex, heterogeneous meshes (e.g., Claude Code spawning OpenClaw specialists), tracking the "Chain of Thought" and "Chain of Command" becomes difficult. Frameworks are starting to introduce their own IDs (e.g., Claude's `agent_id`), but there is no universal standard for cross-framework traceability. The Structured Lineage Broker enforces a standardized identity and mission tracking protocol.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce standardized `x-agent-id` and `x-mission-id` headers for all inter-framework coordination requests.
    * Provide a machine-readable "Lineage Trace" that maps the parent-child relationship of every agent in a mission.
    * Enable real-time observability of the reasoning mesh.
    * Integrate with the Action-Chain Sovereignty Monitor (ACSM) for policy enforcement.
* **Non-Goals:**
    * Replacing framework-specific internal logging.
    * Validating the reasoning content itself (handled by other middlewares).

## 3. Critical User Journey (CUJ)
* **User Persona:** AI Systems Architect
* **Primary Goal:** Trace a failed tool call in an OpenClaw specialist back to the original Claude parent session.
* **The Happy Path (Tasks):**
    1. A Claude parent agent spawns an OpenClaw subagent to perform a specialized database task.
    2. Claude includes its hardware-attested `mission_id` and `agent_id` in the delegation request.
    3. The Structured Lineage Broker intercepts the request and generates a new, linked `agent_id` for the OpenClaw specialist.
    4. Every tool call made by the OpenClaw specialist is tagged with this linked lineage.
    5. When a tool call fails, the architect uses the Mesh Lineage Explorer to view the complete trace, identifying Claude as the parent.

## 4. Design & Architecture
* **System Flow:**
    * **Header Injection**: Automatically injects and propagates lineage headers.
    * **Lineage Registry**: A volatile, high-speed store for active agent relationships.
    * **Trace Aggregator**: Collects events from the A2A hub and coordination bridges.
* **APIs / Interfaces:**
    * `RegisterSubagent(parentAgentId, subagentMetadata) -> subagentId`
    * `GetLineage(agentId) -> agentChain`
    * `VerifyMissionAuthority(missionId, agentId) -> bool`
* **Data Storage/State:** High-performance key-value store for mapping agent hierarchies within a mission.

## 5. Alternatives Considered
* **Distributed Tracing (e.g., OpenTelemetry)**: Valuable for performance, but lacks the "Intent-Aware" and "Attested-Lineage" requirements of secure agent coordination.
* **Log-Based Reconstruction**: Rejected due to latency and the risk of non-deterministic log formats across frameworks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage is hardware-attested, preventing subagents from "Ghosting" or spoofing their parentage.
* **Observability:** Directly powers the Mesh Lineage Explorer UI.

## 7. Evolutionary Changelog
* **2026-03-27:** Initial Document Creation.
