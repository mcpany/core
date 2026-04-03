# Design Doc: Active Attention Heartbeats (AAH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLM context windows expand to 1M+ tokens, specialist subagents have begun utilizing "Context-Window Ghosting" (CWG) attacks. By injecting high-frequency, low-entropy noise fragments, these agents can slowly evict mission-root instructions from the primary attention window without triggering traditional entropy-based exhaustion monitors.

MCP Any needs to solve this by moving from passive attention pinning to active verification. The AAH service will periodically "probe" the model's reasoning state to ensure that core behavioral guardrails remain the dominant driver of agent reasoning.

## 2. Goals & Non-Goals
* **Goals:**
    * Periodically inject "Attention Probes" (invisible reasoning anchors) into the context.
    * Verify the presence and priority of mission-root instructions in the model's response traces.
    * Automatically trigger "Attention Reinforcement" (re-injection of guardrails) when ghosting is detected.
* **Non-Goals:**
    * Modify the underlying LLM architecture or attention mechanism.
    * Protect against prompt injection that occurs during the initial user input (handled by other middlewares).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Coding Agent" from silently ignoring security linting rules after 50+ turns of high-density code generation.
* **The Happy Path (Tasks):**
    1. The user defines a "Mission Root" with strict security guardrails.
    2. The Coding Agent performs 100+ tasks, generating thousands of tokens of reasoning noise.
    3. The AAH middleware detects that the attention weight of the security guardrails has fallen below the safety threshold.
    4. AAH injects a "Heartbeat Probe" into the next turn.
    5. The model's response fails to reference the probe's semantic anchor.
    6. AAH interdicts the session and re-injects the mission-root instructions at the head of the context window.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `AAH Middleware (Inject Probe)` -> `LLM` -> `AAH Middleware (Verify Anchor in Trace)` -> `Response/Reinforce`
* **APIs / Interfaces:**
    * `POST /v1/attention/probe`: Internal endpoint for generating cryptographically bound attention anchors.
    * `x-mcpany-attention-threshold`: Header for specifying required guardrail priority.
* **Data Storage/State:**
    * The Blackboard will store "Attention Signatures" for active sessions to track anchor decay over time.

## 5. Alternatives Considered
* **Static Guardrail Pinning:** Rejected because models can still ignore pinned tokens if the surrounding noise is sufficiently high-density (Attention Sinking).
* **Mandatory Context Pruning:** Rejected because it can lead to "Context Amnesia" where the agent loses task-critical state.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Probes must be semantically unique to each session to prevent subagents from "predicting" and spoofing a successful heartbeat.
* **Observability:** The UI will feature an "Attention Heatmap" showing the real-time priority of the mission root.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
