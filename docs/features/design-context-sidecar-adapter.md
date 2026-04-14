# Design Doc: Context Sidecar Orchestration (CSO) Adapter
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
The stabilization of OpenClaw's `ContextEngine` plugin interface has enabled a "Context-as-a-Service" model. However, swarms using different frameworks (e.g., AutoGen and Claude Code) still suffer from "Context Amnesia" when state is not shared or semantically isolated. The CSO Adapter allows MCP Any to host OpenClaw-compatible context engines as secure sidecars, ensuring that "Mission-Root" state is preserved across framework boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Host OpenClaw-compatible `ContextEngine` plugins natively within MCP Any.
    * Provide a standardized "Context Bus" for inter-framework state synchronization.
    * Ensure "Intent-Bound" isolation for different context fragments.
    * Support "Sovereignty-Aware Compression" to protect mission intents during summarization.
* **Non-Goals:**
    * Implementing a new vector database (it bridges to existing ones).
    * Managing framework-specific reasoning loops (it only manages the state they consume).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Architect
* **Primary Goal:** Share long-term memory between an OpenClaw specialist and a Claude Code orchestrator without losing the "Mission Root" goal.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a mission with a "Mission Root" intent.
    2. CSO Adapter spawns a "Context Sidecar" for this mission.
    3. OpenClaw specialist writes reasoning fragments to the sidecar.
    4. CSO Adapter semantically tags these fragments with the mission intent.
    5. Claude Code orchestrator queries the sidecar for relevant context.
    6. CSO Adapter provides a "Sovereignty-Aware" summary that prioritizes the mission root.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Framework] <-> [Context Bus] <-> [CSO Adapter] <-> [ContextEngine Plugin]`
* **APIs / Interfaces:**
    * `ContextBus`: `PushFragment(fragment Fragment) error`
    * `SovereigntyProvider`: `GetIntentPinnedSummary(missionID string) Summary`
* **Data Storage/State:**
    * Bridges to persistent storage providers defined by the plugin (e.g., Redis, SQLite).

## 5. Alternatives Considered
* **Framework-Specific Bridges:** Rejected as it leads to "N*M" complexity.
* **Centralized Database:** Rejected because it doesn't support the pluggable, distributed nature of modern context strategies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All context fragments are cryptographically bound to the mission root.
* **Observability:** State synchronization latency and "Intent Consistency" scores are visualized in the dashboard.

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
