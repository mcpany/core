# Design Doc: ContextEngine Plugin Adapter
**Status:** Draft
**Created:** 2026-04-25

## 1. Context and Scope
The release of OpenClaw v2026.3.7-beta.1's "ContextEngine" has introduced a standardized, pluggable interface for context management. MCP Any needs a native adapter to host these plugins, allowing agents from disparate frameworks to utilize specialized state strategies (e.g., "Cognitive Anchoring," "Intent-Aware Compression") while maintaining a centralized security and audit boundary.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a native host for OpenClaw-compatible ContextEngine plugins.
    * Provide "Cognitive Anchoring" to protect critical mission intents from "Context-Splicing."
    * Standardize state mutation signals across connected agent frameworks.
* **Non-Goals:**
    * Replacing existing framework-specific state management (e.g., AutoGen's internal state).
    * Providing a standalone database (will leverage the Shared KV Store/Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agentic Software Engineer
* **Primary Goal:** Implement a custom "Long-Term Memory" strategy for a swarm without modifying the agent's core reasoning logic.
* **The Happy Path (Tasks):**
    1. Engineer develops an OpenClaw-compatible `ContextEngine` plugin.
    2. Engineer registers the plugin with MCP Any via the `config.yaml`.
    3. MCP Any's ContextEngine Plugin Adapter loads the plugin and binds it to the specified agent mission.
    4. During reasoning, the agent's state mutation requests are intercepted and processed by the custom plugin.
    5. The adapter ensures the resulting state conforms to the "Cognitive Anchoring" policy before committing to the Blackboard.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [Context Bridge] -> [ContextEngine Plugin Adapter] -> [Shared KV Store (Blackboard)]`
* **APIs / Interfaces:**
    * `ContextEngine.onContextUpdate()`: Hook for processing context mutations.
    * `ContextEngine.getAnchoredIntent()`: Interface for retrieving the protected root mission intent.
* **Data Storage/State:**
    * Plugin-specific micro-state is stored in the Shared KV Store, isolated by `mission_id`.

## 5. Alternatives Considered
* **Internal State Forking:** Rejected as it lacks interoperability with the growing OpenClaw ecosystem.
* **Hardcoded Summarization Logic:** Rejected due to lack of flexibility for specialized agent domains (e.g., legal or medical reasoning).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Plugins are executed in an isolated WASM sandbox (via the WASM-BSH Sanitizer) to prevent unauthorized host access.
* **Observability:** State mutations are logged in the Mission Audit Trail, including the ID of the plugin that performed the mutation.

## 7. Evolutionary Changelog
* **2026-04-25:** Initial Document Creation.
* **2026-05-14:** Evolving to support OpenClaw v2026.3.7 "ContextEngine" lifecycle hooks. This update enables MCP Any to act as a universal host for pluggable context plugins, neutralizing "Context Amnesia" in deep, heterogeneous swarms.
* **2026-05-15:** Introducing "Intent-Bound Memory Isolation." This evolution ensures that "Mission-Root" anchors are cryptographically protected and semantically isolated within the plugin host, preventing "Context Ghosting" where critical goals are discarded during automated state compression or summarization.
<<<<<<< HEAD
* **2026-06-28:** Implementing "Multi-Tenant Context Isolation." This update leverages the matured OpenClaw v2026.3.7 lifecycle hooks to enforce strict state separation between disparate agent missions. It ensures that pluggable context strategies cannot "leak" state or reasoning fragments across framework boundaries, maintaining absolute cognitive sovereignty in shared execution environments.

### Update: 2026-03-26 - Action-Chain Governance Integration
**Context:** Today's market sync revealed the rise of "Insider Threat" agents and machine-speed "Swarm Attacks" (GTG-1002).
**Architecture Adjustment:**
* Integrating the **Action-Chain Sovereignty Monitor (ACSM)** into the ContextEngine lifecycle.
* Mandatory validation of state-transition sequences against the mission-root manifest before plugin-mediated context updates.
**Security Impact:** Prevents malicious subagents from using pluggable context hooks to "chain" unauthorized system actions at machine speed.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
