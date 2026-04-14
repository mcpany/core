# Design Doc: Adaptive Thinking Proxy
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
With the release of Claude Opus 4.6's "Adaptive Thinking" and Gemini's `x-gemini-reasoning-effort`, model providers are increasingly exposing controls that allow developers to balance model intelligence (reasoning depth) against speed and cost. However, these controls are fragmented, with varying headers and parameter names.

MCP Any needs to provide a unified "Intelligence-Cost Proxy" that abstracts these provider-specific effort controls. This allows agent frameworks and swarms to define a generic "Thought Budget" that MCP Any then translates and enforces across any connected LLM backend.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified API/header for controlling reasoning effort (`x-mcpany-reasoning-effort`).
    * Map MCP Any effort levels (e.g., `low`, `medium`, `high`) to provider-specific parameters (Claude's `effort`, Gemini's `reasoning_effort`).
    * Implement hardware-attested "Thought Budgets" to prevent subagents from exceeding parental cost constraints.
    * Provide real-time telemetry on the relationship between reasoning effort and actual token consumption.
* **Non-Goals:**
    * Automatically adjusting effort based on task complexity (this is the agent's or model's responsibility).
    * Supporting model-specific reasoning features that cannot be generalized (e.g., specific Gemini "thinking" tokens).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Infrastructure Architect
* **Primary Goal:** Limit the total cost of a multi-agent "Search & Code" mission by capping the reasoning effort of specialist subagents.
* **The Happy Path (Tasks):**
    1. The Architect configures a mission-root policy in MCP Any with a `max_reasoning_effort: medium`.
    2. A "Lead Agent" spawns 3 "Specialist Teammates" to perform parallel research.
    3. Specialist A attempts to call Claude with `effort: high` to solve a complex bug.
    4. MCP Any's Adaptive Thinking Proxy intercepts the request, detects the mission-root constraint, and automatically downgrades the request to `effort: medium`.
    5. The Proxy adds an `x-mcpany-effort-capped: true` header to the response to inform the Lead Agent of the constraint enforcement.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCP Any: Request (x-mcpany-reasoning-effort: high)
        MCP Any->>Policy Engine: Check Mission-Root Budget
        Policy Engine-->>MCP Any: Allow (max: medium)
        MCP Any->>Adaptive Thinking Proxy: Translate & Cap
        Adaptive Thinking Proxy->>Claude: Request (effort: medium)
        Claude-->>Adaptive Thinking Proxy: Response
        Adaptive Thinking Proxy-->>Agent: Response (x-mcpany-effort-capped: true)
    ```
* **APIs / Interfaces:**
    * New Header: `x-mcpany-reasoning-effort: low | medium | high | extreme`
    * Configuration: `thought_budget: { max_effort: string, max_tokens_per_effort_unit: int }`
* **Data Storage/State:**
    * State is managed within the mission-root context sidecar, tracking cumulative "Reasoning Units" consumed across the swarm.

## 5. Alternatives Considered
* **Direct Pass-through:** Rejecting unified headers and forcing agents to know provider-specific effort keys. *Rejected* because it breaks framework-neutrality, a core pillar of MCP Any.
* **Hard Token Caps Only:** Using only token limits instead of effort levels. *Rejected* because token limits don't prevent the model from "thinking" too hard on a single turn, which impacts latency even if token usage is within bounds.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Reasoning effort can be used as a DoS vector (Reasoning Entropy Exhaustion). The Proxy will integrate with the CLS Controller to drop requests that exceed mission-wide "Cognitive Loads."
* **Observability:** Metrics will be exported showing `effort_requested` vs `effort_delivered` vs `tokens_consumed`, visualized in the UI Roadmap's "Reasoning-Effort Control Dashboard."

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
