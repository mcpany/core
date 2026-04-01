# Design Doc: Agentic Entropy Monitor (AEM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from linear execution to high-density, horizontal meshes, maintaining "Cognitive Coherence" has become a critical challenge. Specialist subagents often generate high-entropy reasoning traces that can "pollute" the attention window of the mission-root agent, leading to instruction eviction and mission drift. This phenomenon, known as "Semantic Entropy Leakage," threatens the stability of long-running autonomous missions.

The Agentic Entropy Monitor (AEM) is designed to act as a kernel-level coherence broker within MCP Any. It provides real-time analysis of the semantic stability of subagent outputs, ensuring that "Noise" is interdicted before it compromises the mission-root sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, sub-millisecond semantic entropy scoring for all inter-agent coordination fragments.
    * Trigger automated "Re-alignment Quorums" when entropy exceeds pre-defined mission-root thresholds.
    * Provide hardware-attested interdiction of high-entropy "Noise" fragments.
    * Support configurable "Entropy Profiles" for different specialist roles.
* **Non-Goals:**
    * Modifying the internal weights or parameters of the connected LLMs.
    * Replacing the primary reasoning engine's own self-correction loops (AEM acts as an external governor).
    * Providing natural language explanations for entropy scores (AEM is a high-speed security primitive).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Prevent a "Thinking Tool" or specialist subagent from evicting core security guardrails in a 1M+ token context window.
* **The Happy Path (Tasks):**
    1. The Architect defines an "Entropy Budget" for a specialist subagent (e.g., "Max Entropy: 0.85").
    2. The specialist subagent begins a high-intensity "Chain-of-Thought" expansion.
    3. The AEM intercepts every reasoning fragment on the coordination bus.
    4. AEM calculates the "Instruction-Path Entropy" (IPE) score for each fragment.
    5. A fragment exceeds the 0.85 threshold.
    6. AEM automatically routes the fragment to a "Paraphrasing Sandbox" for normalization.
    7. The normalized, low-entropy fragment is delivered to the mission-root agent.
    8. The mission-root retains its core instructions (attention anchors) while absorbing the sub-task result.

## 4. Design & Architecture
* **System Flow:**
    `Subagent Output` -> `AEM Hook (Interception)` -> `Semantic Entropy Engine (IPE Calculation)` -> `Threshold Evaluator` -> `Interdiction Logic (Block/Normalize/Pass)` -> `Shared Mailbox/Blackboard`.
* **APIs / Interfaces:**
    * `POST /v1/entropy/score`: Submit a text/modal fragment for real-time scoring.
    * `PUT /v1/missions/{id}/entropy-policy`: Configure thresholds and actions for a specific mission branch.
    * `x-mcpany-entropy-score`: Header injected into all inter-agent messages.
* **Data Storage/State:**
    * Real-time entropy scores are stored in a high-speed, in-memory CRDT shard to support parallel teammate coordination.
    * Historical coherence metrics are persisted to the SQLite "Blackboard" for RL-telemetry export.

## 5. Alternatives Considered
* **Passive Logging:** Rejected because it allows "Attention Drift" to occur before a human or auditor can react.
* **Global Token Budgeting:** Rejected because it doesn't distinguish between high-utility "Complex Thought" and high-entropy "Noise."
* **Inside-the-Model Guardrails:** Rejected because different models have varying levels of self-awareness; infrastructure-level governance provides a universal baseline.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AEM must be hardware-attested (TPM/SE) to prevent subagents from "spoofing" their own entropy scores.
* **Observability:** High-entropy alerts are streamed to the "Agentic Entropy Scoreboard" in the UI for real-time architect monitoring.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
