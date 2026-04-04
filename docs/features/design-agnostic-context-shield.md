# Design Doc: Agnostic Context Shield (ACS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Conventional semantic guards in the Universal Agent Bus focus on "Intent Drift" or "Instruction Splicing." However, the discovery of Intent-Agnostic Context Poisoning (IACP) reveals a gap: attackers can inject low-entropy, "intent-neutral" metadata into shared shards that biases model weights during reasoning without triggering drift alerts.

The ACS is required to monitor the "noise floor" of the context window. It analyzes shared shards for long-range reasoning influence patterns that are semantically benign on their own but collectively malicious.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect low-entropy context fragments that exhibit high "Influence Scores" on agent reasoning.
    * Sanitize shared shards of "intent-neutral" metadata that violates local behavioral guardrails.
    * Provide a "Semantic Noise Floor" baseline for all mission branches.
* **Non-Goals:**
    * Blocking legitimate coordination metadata.
    * Performing full chain-of-thought analysis (ACS is a pre-reasoning guard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor for Agent Swarms
* **Primary Goal:** Prevent "Silent Bias" in a swarm of 50 specialized subagents sharing a large context window.
* **The Happy Path (Tasks):**
    1. ACS establishes a "Behavioral Baseline" for the mission-root.
    2. Specialist agents write to the shared teammate mailbox shards.
    3. ACS performs real-time "Influence Analysis" on fragments, even those marked as "Metadata."
    4. Fragments that attempt to bias reasoning toward restricted toolsets (e.g., unauthorized shell usage) are quarantined.
    5. Auditor reviews the "Silent Bias" alert in the dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Write] --> B[Blackboard/Mailbox]
        B --> C{ACS Middleware}
        C -->|Low Influence| D[Shard Persistent]
        C -->|High Bias Potential| E[Quarantine/Redact]
        F[Mission Root Baseline] --> C
    ```
* **APIs / Interfaces:**
    * `acs.AnalyzeFragment(fragment_id)`: Internal call during state writes.
    * `/v1/acs/baseline`: Configures the expected semantic distribution for a mission.
* **Data Storage/State:**
    * Vector-based "Influence Maps" stored in the local episodic cache.

## 5. Alternatives Considered
* **Strict Schema Enforcement:** Rejected as IACP often uses valid but biased natural language.
* **Instruction-only Sanitization:** Rejected as it misses the "Agnostic" nature of the metadata attack.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ACS operates at the kernel level of the State Broker (ZCMB/UEG).
* **Observability:** Bias scores are visualized in the "Attention Heatmap" UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
