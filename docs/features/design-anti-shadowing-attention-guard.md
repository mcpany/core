# Design Doc: Anti-Shadowing Attention Guard
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the shift toward 1M+ token context windows and dynamic attention management in large language models, a new exploit pattern has emerged: **Attention Shadowing**. Malicious or poorly-designed subagents can flood the context with high-frequency, low-entropy fragments (e.g., repeating status messages or irrelevant reasoning steps).

Models that use frequency or "recency of activity" as priority signals for their attention mechanisms may displace critical **Mission-Root Anchors** (instructions, security policies, core goals) in favor of this noise. The **Anti-Shadowing Attention Guard** extends the ALRA framework to ensure that anchors remain permanent regardless of subagent noise patterns.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Frequency-Aware Intent Pinning" (FAIP).
    * Detect and block high-frequency, low-entropy noise patterns from subagents.
    * Provide a "Shadowing Risk Score" for every reasoning fragment.
    * Ensure Mission-Root anchors have an immutable attention priority.
* **Non-Goals:**
    * Modifying the model's internal transformer weights.
    * Reducing the overall token limit (focus is on *what* occupies the window).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Protect core safety guardrails from being "pushed out" of the model's attention by a verbose specialist agent.
* **The Happy Path (Tasks):**
    1. The Mission-Root sets a "Permanent Safety Anchor" via MCP Any.
    2. A specialist subagent starts emitting high-frequency reasoning fragments (10/sec).
    3. The Anti-Shadowing Attention Guard monitors the entropy and frequency of these fragments.
    4. The Guard identifies a shadowing risk.
    5. The Guard dynamically injects "Attention Reinforcement" tokens to re-assert the Safety Anchor's priority.
    6. The Guard throttles the specialist subagent's message frequency.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        SA[Subagent] -- Fragments --> AG[Attention Guard]
        AG -- Analyzes Frequency/Entropy --> PM[Priority Manager]
        PM -- Reinforces --> MRA[Mission Root Anchor]
        AG -- Throttles --> SA
        MRA -- Pinned In --> CW[Context Window]
    ```
* **APIs / Interfaces:**
    * `x-mcpany-attention-priority`: Header for defining anchor weight.
    * `x-mcpany-shadow-limit`: Header for setting frequency thresholds.
* **Data Storage/State:**
    * Real-time frequency counters stored in high-speed Redis or in-memory caches.
    * Stylometric baselines for subagents to distinguish between "productive reasoning" and "noise".

## 5. Alternatives Considered
* **Static Context Truncation:** Simply cutting off subagent history. *Rejected* because it may break legitimate complex reasoning paths.
* **Global Rate Limiting:** Throttling all subagents equally. *Rejected* because it reduces swarm efficiency for tasks that actually require high-frequency coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Guard itself must be hardware-attested to prevent subagents from modifying its thresholds.
* **Observability:** Real-time heatmap of context window "Attention Density" in the Visual Attention Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
