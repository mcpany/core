# Design Doc: Adaptive Resource Reclamation (ARR) Service
**Status:** Draft
**Created:** 2026-07-23

## 1. Context and Scope
As AI agent swarms move toward high-density parallel execution, static resource limits (tokens, reasoning time) create significant performance bottlenecks. Specialist agents often encounter mission-critical reasoning bursts that exceed their initial budgets, while low-priority teammates may have excess, unused capacity. This "Resource Squatting" leads to unnecessary cognitive stalls and mission failures.

The Adaptive Resource Reclamation (ARR) Service provides a mesh-wide mechanism to dynamically re-allocate reasoning and token budgets between teammates based on real-time task urgency and mission priority.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time monitoring of subagent budget utilization.
    * Provide a standardized "Urgency Signal" for agents to request budget expansion.
    * Support autonomous, mission-root authorized budget re-allocation between teammates.
    * Ensure hardware-attested transparency for all budget shifts.
* **Non-Goals:**
    * Directly managing LLM provider quotas (this sits at the infrastructure/billing layer).
    * Predicting future reasoning needs without explicit agent signals.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure a high-stakes "Security Specialist" agent has enough tokens to complete a deep exploit analysis by reclaiming budget from a background "Documentation Specialist."
* **The Happy Path (Tasks):**
    1. The Security Specialist detects a complex code pattern and issues a "High Urgency" reasoning signal to MCP Any.
    2. The ARR Service scans the teammate mesh and identifies the Documentation Specialist as having 80% unused token budget and low priority.
    3. The ARR Service issues a hardware-locked "Budget Recall" to the Documentation Specialist.
    4. The ARR Service issues an "Attested Budget Grant" to the Security Specialist.
    5. Both agents resume reasoning with updated constraints, and the mission-root blackboard records the attribution shift.

## 4. Design & Architecture
* **System Flow:**
    * [Agent] -> (Urgency Signal) -> [ARR Service]
    * [ARR Service] -> (Inventory Check) -> [Mesh Budget Registry]
    * [ARR Service] -> (Recall/Grant) -> [Subagents]
* **APIs / Interfaces:**
    * `POST /v1/mesh/reclaim`: Internal endpoint for the ARR to trigger a re-allocation cycle.
    * `x-mcpany-urgency`: New header for agents to signal reasoning pressure.
* **Data Storage/State:**
    * State is managed in the **Mission-Root Budget Registry**, which is a hardware-attested sidecar to the Blackboard.

## 5. Alternatives Considered
* **Static Over-provisioning**: Rejected due to high token costs and increased risk of "Spiral of Death" loops.
* **Manual HITL Budgeting**: Rejected due to prohibitive MTTC (Mean Time to Coordinate) in high-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budget shifts must be signed by the ARR Service and validated against the Mission-Root manifest to prevent "Economic Hijacking."
* **Observability:** Track reclamation frequency and "Urgency Accuracy" to optimize future mission-root allocations.

## 7. Evolutionary Changelog
* **2026-07-23:** Initial Document Creation.
