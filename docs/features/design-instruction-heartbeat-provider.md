# Design Doc: Instruction Heartbeat Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLM context windows scale to 1M+ tokens, model providers have implemented aggressive "Context-Window Garbage Collection" (CWGC) to manage KV-cache efficiency. However, this has introduced "GC Fragility," where critical behavioral guardrails and "Mission-Root" instructions—often located at the beginning of the prompt—are evicted or deprioritized by the attention mechanism in favor of newer, high-entropy subagent logs.

The Instruction Heartbeat Provider (also known as "Sentinel") is a cognitive stability service that programmatically reinforces mission-critical instructions. It works by dynamically injecting hardware-attested "Instruction Heartbeats" into the attention layer at regular intervals, ensuring that the model's reasoning remains anchored to the primary intent despite aggressive token cleanup.

## 2. Goals & Non-Goals
* **Goals:**
    * Programmatically reinforce mission-root behavioral guardrails.
    * Neutralize "GC Fragility" by ensuring core instructions remain "GC-Immune."
    * Provide hardware-attested proof that "Instruction Heartbeats" were successfully ingested by the model.
    * Integrate with Attention-Locked Reasoning Anchors (ALRA) for multi-layered attention governance.
* **Non-Goals:**
    * Modifying the model's base training or fine-tuning.
    * Replacing context summarization (it complements it by protecting the "un-summarizable" root intent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Maintain a "Do not access PII" guardrail throughout a multi-day autonomous research mission.
* **The Happy Path (Tasks):**
    1. The mission-root is established with a high-priority "PII Access: DENY" instruction.
    2. The Instruction Heartbeat Provider monitors the LLM's attention-utilization and KV-cache pressure.
    3. As subagents generate thousands of research tokens, the Sentinel detects that the "PII Access" guardrail is at risk of eviction.
    4. The Sentinel injects a TPM-signed "Instruction Heartbeat" (a semantically equivalent but fresh fragment) into the latest coordination turn.
    5. The model's attention window is "refreshed," and the guardrail remains active and prioritized.
    6. The mission concludes without any PII leaks, despite massive context turnover.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Monitor[Attention Monitor] -->|KV-Cache Pressure| Sentinel[Instruction Heartbeat Provider]
        Root[Mission Root Store] -->|Guardrail Template| Sentinel
        Sentinel -->|Sign Fragment| TPM[Hardware TPM]
        TPM -->|Attested Heartbeat| Injector[Context Injector]
        Injector -->|Inference Turn| LLM[LLM Reasoning Engine]
    ```
* **APIs / Interfaces:**
    * `Middleware: Sentinel_Heartbeat_Monitor`: Background analyzer for attention drift.
    * `Policy: heartbeat_frequency`: Defines how often heartbeats are injected (e.g., every 10k tokens or every 5 turns).
* **Data Storage/State:**
    * "Heartbeat Templates" are stored in the secure Mission Root segment of the Blackboard.

## 5. Alternatives Considered
* **Static Pinning (KV-Cache Locking):** Often insufficient if the model provider's API doesn't support explicit cache-locking for the entire session. Heartbeats provide a "client-side" guarantee of attention presence.
* **Instruction Prefixing (on every message):** Rejected due to high token cost and "Reasoning Exhaustion."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Heartbeats must be hardware-attested to prevent subagents from injecting "Fake Heartbeats" that override parent intents.
* **Observability:** Heartbeat events are visualized on the "Visual Attention Dashboard" as "Reinforcement Pulses."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
