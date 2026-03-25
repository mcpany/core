# Design Doc: Active Attention Steering (AAS) Adapter
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
As AI agents move into 2M+ token context windows, the risk of "Context Drift" and "Attention Eviction" has reached a critical threshold. High-volume tool outputs, logs, and multimodal metadata can "push" the original mission-root instructions out of the model's active attention layer, leading to specialist subagents diverging from their parent's goals.

MCP Any needs to solve this by evolving from a passive context bridge to an active cognitive governor. The Active Attention Steering (AAS) Adapter will ensure that the hardware-attested mission root remains the primary driver of all subagent reasoning, regardless of the volume of secondary data.

## 2. Goals & Non-Goals
* **Goals:**
    * Force-inject hardware-attested mission-root anchors into every subagent reasoning turn.
    * Prevent "Attention Eviction" by prioritizing core instructions in the attention window.
    * Provide cryptographic proof that a subagent was "steered" by the mission root.
    * Dynamically adjust steering intensity based on detected reasoning entropy.
* **Non-Goals:**
    * Modifying the underlying LLM's attention mechanism (this is a middleware implementation).
    * Restricting the subagent's ability to ingest tool data (only ensures instructions stay pinned).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Architect / Enterprise Security Admin
* **Primary Goal:** Ensure a "Refactoring Agent" stays bound to secure coding standards even when processing 500,000 lines of legacy code.
* **The Happy Path (Tasks):**
    1. Architect defines a "Mission Root" with hardware-attested security constraints.
    2. refactoring subagent begins ingesting large volumes of code.
    3. AAS Adapter monitors the subagent's reasoning entropy.
    4. As the context window fills, AAS force-injects "Attention Anchors" into the next reasoning prompt.
    5. Subagent's next reasoning turn correctly references the original security constraints.
    6. System logs a "Steered Turn" event with hardware-bound attestation.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Request] -> [Reasoning Entropy Monitor] -> [AAS Injector] -> [Upstream LLM]
    AAS Injector retrieves the hardware-attested mission root from the Mission Registry and appends it to the "User" or "System" block of the outgoing request, utilizing Gemini v0.45.0 `x-aas-steer` headers where available.
* **APIs / Interfaces:**
    * `POST /v1/steering/anchor`: Register a mission-root anchor.
    * `GET /v1/steering/status`: Monitor attention-steering metrics for a session.
* **Data Storage/State:**
    Mission-root anchors are stored in the Hardware-Attested Mission Registry (HAMM) and keyed by session ID.

## 5. Alternatives Considered
* **Passive Pinnning:** Simply placing instructions at the top/bottom of the prompt. Rejected because large-context models still suffer from "Lost in the Middle" or eviction when log volume is high.
* **Token Pruning:** Removing tool outputs to save space. Rejected because subagents often need the full context for correct reasoning; steering is safer than pruning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Steering anchors must be cryptographically signed. Unauthorized agents cannot inject their own anchors.
* **Observability:** Real-time "Attention Map" visualization in the UI, showing the ratio of mission-root attention vs. tool-output entropy.

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
