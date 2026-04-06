# Design Doc: Entropy-Aware Resource Controller (EARC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Deep agent swarms are susceptible to "Cognitive Meltdowns"—hallucinatory loops where specialist subagents begin to generate high-entropy noise that diverges from the mission-root intent. Current resource controllers (based on token count or time) fail to detect the *quality* of the reasoning drift until significant budget has been exhausted.

The Entropy-Aware Resource Controller (EARC) implements real-time statistical analysis of reasoning traces. It identifies patterns of increasing semantic entropy and automatically triggers "Reasoning Resets" or budget revocation before mission-critical guardrails are compromised.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect hallucinatory loops in sub-millisecond intervals.
    * Automatically throttle or terminate "Entropy-High" reasoning branches.
    * Provide a standardized "Coherence Score" for all agent outputs.
    * Integrate with the Reasoning-Budget Firewall (RBF) for dynamic quota adjustment.
* **Non-Goals:**
    * Modifying the LLM generation process itself (EARC is a monitoring and interdiction layer).
    * Providing natural language explanations for entropy spikes (it is a statistical gate).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Auditor
* **Primary Goal:** Prevent a specialist subagent from exhausting $50 of API credits on a recursive "Thank you" loop.
* **The Happy Path (Tasks):**
    1. Specialist Agent enters a refinement loop.
    2. EARC monitors the "Statistical Entropy" of each turn.
    3. Entropy exceeds the `mission-root` threshold for 3 consecutive turns.
    4. EARC issues an `INTERDICT` signal to the Gateway.
    5. The session is terminated, and a "Coherence Alert" is sent to the UI.

## 4. Design & Architecture
* **System Flow:**
    [Reasoning Fragment] -> [EARC Entropy Engine] -> [Coherence Scoring] -> [Policy Check] -> [Interdiction/Allow]
* **APIs / Interfaces:**
    * `GET /v1/metrics/coherence`: Retrieve current coherence scores for active agents.
    * `POST /v1/governance/interdict`: Manually trigger interdiction based on UI alerts.
* **Data Storage/State:**
    * Rolling buffers of semantic vectors for active reasoning traces.
    * Threshold manifests bound to mission-root security policies.

## 5. Alternatives Considered
* **Pattern-Based Filtering:** Rejected due to inability to detect novel hallucinatory patterns.
* **Manual HITL Review:** Rejected due to unacceptable latency in autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** EARC uses hardware-attested reasoning traces to ensure the monitor itself isn't bypassed.
* **Observability:** Entropy Scoreboard in the UI provides a real-time visualization of swarm "sanity" levels.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
