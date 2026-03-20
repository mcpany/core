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
