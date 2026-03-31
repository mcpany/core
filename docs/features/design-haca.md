# Design Doc: Hardware-Attested Cost Attribution (HACA)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As AI agent swarms evolve from experimental scripts to production enterprise meshes, the "Resource Sovereignty" of the mission root becomes critical. Modern models (like those in Gemini CLI) are shifting toward tiered subscription and effort-based quotas. Without a standardized way to attribute costs and reasoning effort across a deep, multi-framework swarm, organizations face "Economic Drift" where rogue subagents or inefficient refinement loops exhaust budgets without accountability.

MCP Any needs to provide a hardware-locked mechanism to track and attribute every token consumed and every compute millisecond spent to its specific mission-root lineage. This ensures that budgets are respected across framework boundaries (Claude, OpenClaw, AutoGen) and that resource consumption is non-repudiable.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically attribute every tool call and reasoning fragment to its mission-root lineage.
    * Enforce hardware-locked reasoning budgets (ARE) and token quotas.
    * Provide a non-repudiable audit trail for economic accountability in multi-cloud swarms.
    * Support real-time budget reconciliation during cross-framework handoffs.
* **Non-Goals:**
    * Implementing a billing/payment gateway (this attributes cost, doesn't process payments).
    * Optimizing the actual cost of LLM calls (that is the job of context compactors).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Platform Ops Manager
* **Primary Goal:** Ensure that a "Market Research" swarm led by a Claude agent doesn't exceed its $50 budget when delegating sub-tasks to 10+ OpenClaw specialist agents.
* **The Happy Path (Tasks):**
    1. User initiates a mission and defines a TPM-signed budget manifest (Token Quota + Reasoning Effort limit).
    2. MCP Any issues a mission-root identity token embedded with these constraints.
    3. The primary agent (Claude) delegates a "Web Search" task to an OpenClaw subagent via the A2A bridge.
    4. MCP Any intercepts the handshake, validating that the subagent inherits the cost-attribution headers and remains within the mission-root budget.
    5. The subagent executes tool calls; MCP Any cryptographically signs each request with the mission-root ID and decrements the central budget.
    6. If the subagent attempts a tool call that would exceed the remaining budget, MCP Any interdicts the call and alerts the mission-root.

## 4. Design & Architecture
* **System Flow:**
    [Mission Root] -> (Signed Budget Manifest) -> [MCP Any HACA Provider]
    [MCP Any HACA Provider] -> (Lineage-Bound Headers) -> [Subagent A]
    [Subagent A] -> (Tool Call + Lineage Proof) -> [MCP Any Gateway] -> [Upstream LLM/MCP Server]
* **APIs / Interfaces:**
    * `GET /v1/haca/budget/:mission_id`: Retrieve real-time budget status.
    * `POST /v1/haca/attribute`: Internal hook for middleware to log consumption.
* **Data Storage/State:**
    * State is managed in a secure, hardware-bound SQLite vault, with mission-root signatures required for budget increments.

## 5. Alternatives Considered
* **Framework-level Attribution**: Rejected because it cannot enforce limits across frameworks (e.g., Claude cannot control OpenClaw's local SQLite consumption).
* **Network-level Logging**: Rejected because it lacks the "Reasoning Lineage" context; it sees traffic but not "why" the traffic was generated or which mission-root authorized it.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attribution tokens are hardware-bound (TPM) to prevent subagents from "stealing" budget fragments from siblings.
* **Observability:** Real-time dashboards showing "Cost per Intent" and "Effort per Task."

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
