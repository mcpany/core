# Design Doc: Intent-Minimization Proxy (IMP) Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale horizontally across distributed nodes, the volume of context being transferred during handoffs has become a major performance bottleneck (increasing MTTC). Existing solutions like OpenClaw's IMP provide semantic pruning to reduce token overhead. However, aggressive minimization often leads to "Minimization Amnesia," where subagents lose track of critical mission-root guardrails because they are deemed semantically distant from the immediate sub-task.

MCP Any needs an authoritative IMP Middleware that not only optimizes context for transport but also enforces the persistence of "Invariable Instruction Anchors" (IIA) to maintain mission integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic pruning of subagent context fragments to reduce transmission latency.
    * Enforce "Invariable Instruction Anchors" (IIA) to prevent mission-root guardrail erasure.
    * Provide a standardized interface for pluggable minimization strategies (e.g., LLM-based summarization vs. embeddings-based pruning).
    * Synchronize minimized state across distributed AMT tunnels.
* **Non-Goals:**
    * Replacing the primary context window management of the target LLM.
    * Providing long-term episodic memory (handled by the UEG Memory Broker).
    * Modifying the core reasoning traces of the agent.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Sub-second handoff of complex task state between 5 specialists without losing security constraints.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a task to a remote specialist over an AMT tunnel.
    2. IMP Middleware intercepts the context transfer request.
    3. The Proxy identifies "Invariable Instruction Anchors" (marked by the mission root).
    4. The Proxy applies the active minimization strategy to the remaining context fragments (e.g., removing stale tool outputs).
    5. The minimized context + IIA fragments are cryptographically signed and transmitted.
    6. The remote specialist receives a compact, high-utility context that remains anchored to the mission-root guardrails.
    7. Specialist executes the task and returns a minimized result.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Raw Context Data] --> B[Anchor Identifier]
        B --> C{Is Invariable?}
        C -- Yes --> D[IIA Protected Buffer]
        C -- No --> E[Minimization Engine]
        E --> F[Semantic Pruning/Summarization]
        D --> G[Minimized Context Assembler]
        F --> G
        G --> H[Signed Transport Bundle]
    ```
* **APIs / Interfaces:**
    * `imp.MinimizeContext(sessionID, contextFragments) -> MinimizedBundle`: Main entry point for context optimization.
    * `imp.RegisterAnchor(sessionID, fragmentID)`: Marks a specific fragment as "Invariable."
    * `imp.SetStrategy(strategyType, parameters)`: Configures the pruning logic (e.g., threshold-based).
* **Data Storage/State:**
    * **Anchor Registry:** Per-session tracking of IIA fragments.
    * **Minimization Cache:** Temporary storage of previously optimized bundles to reduce redundant processing.

## 5. Alternatives Considered
* **Client-Side Summarization:** Rejected because it places too much compute load on the specialist agents and doesn't provide a centralized enforcement point for security anchors.
* **Static Context Truncation:** Rejected because it is intent-agnostic and frequently removes the most relevant "Instruction" fragments while keeping "Data" fragments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All minimized bundles are cryptographically signed by the IMP to prevent "Context Injection" during the optimization phase.
* **Observability:** Integrated with the "Context Shifting Timeline" in the UI to visualize pruning efficiency and anchor preservation.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
