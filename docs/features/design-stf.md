# Design Doc: Speculative Token Filter (STF)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of high-performance agent reasoning, "Speculative Decoding" has become a standard optimization. However, research on 2026-07-25 revealed the "Speculative Decoding Leakage" (SDL) vulnerability. Attacker-controlled subagents can probe mission-root constraints by analyzing predicted tokens that are subsequently rejected by security middleware but still observable via side-channels (e.g., latency or error messages). MCP Any needs to provide a secure perimeter that filters these speculative tokens before they are exposed to potentially hostile sub-components.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept speculative token streams from reasoning engines.
    * Perform real-time, hardware-locked filtering against mission-root security policies.
    * Ensure that rejected speculative tokens never leave the secure execution perimeter.
    * Neutralize side-channel analysis of rejected predictions.
* **Non-Goals:**
    * Implementing the speculative decoding engine itself (it works as a proxy/filter for existing engines).
    * Optimizing the raw performance of LLM inference.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect for an Agent Swarm.
* **Primary Goal:** Protect sensitive mission-root constraints from being probed by specialist subagents during high-speed reasoning.
* **The Happy Path (Tasks):**
    1. Architect enables STF in the mission-root manifest.
    2. Primary agent starts a task requiring speculative decoding for speed.
    3. Speculative engine predicts a sequence of tokens, some of which contain sensitive mission identifiers.
    4. STF intercepts the predicted sequence.
    5. STF identifies that the sensitive tokens violate the "Redaction Policy."
    6. STF "zeroes out" or masks the sensitive speculative tokens before they are transmitted to the subagent's attention window.
    7. Subagent only receives the sanitized reasoning path, neutralized against SDL.

## 4. Design & Architecture
* **System Flow:**
    `LLM Speculator` -> `STF Middleware (Hardware-Locked)` -> `Security Policy Engine` -> `Sanitized Stream` -> `Subagent`
* **APIs / Interfaces:**
    * `TokenFilter`: `Filter(tokens []Token, policy Policy) ([]Token, error)`
* **Data Storage/State:**
    * Transient hardware-locked buffers for token sequences.

## 5. Alternatives Considered
* **Disabling Speculative Decoding**: Rejected due to the 40%+ performance penalty in deep agent chains.
* **Software-only Sanitization**: Rejected because it remains vulnerable to timing side-channels; hardware-locked filtering provides the necessary isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Extends the security perimeter into the model's predictive buffer.
* **Observability**: Logs the frequency and content of rejected speculative tokens for anomaly detection.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
