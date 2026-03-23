# Design Doc: Negative Feedback Attestation (NFA) Middleware
**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
Autonomous agents frequently suffer from "Stubborn Agent" syndrome, where they ignore or rationalize away negative feedback from users in favor of their internal reasoning loops. Negative Feedback Attestation (NFA) ensures that user corrections are cryptographically bound to the subsequent reasoning trace, making it impossible for the agent to bypass the correction without explicit attestation failure.

## 2. Goals & Non-Goals
* **Goals:**
    * Create a mandatory "Correction Sink" for user feedback.
    * Cryptographically bind NFA tokens to the next tool request.
    * Provide a "stubbornness" audit trail.
* **Non-Goals:**
    * Natural Language Processing of feedback (handled by the LLM).
    * Auto-correction of agent reasoning.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using Claude Code
* **Primary Goal:** Force an agent to stop repeating a failed file edit.
* **The Happy Path (Tasks):**
    1. User submits a correction: "Stop using that library."
    2. MCP Any NFA Middleware generates a signed Correction Token.
    3. Agent's next reasoning cycle must include the NFA token in its header.
    4. Policy engine rejects any tool call that doesn't semantically align with the NFA token.

## 4. Design & Architecture
* **System Flow:**
    `[User Feedback] -> [NFA Tokenizer] -> [Reasoning Context] -> [Tool Validation]`
* **APIs / Interfaces:**
    * `/v1/nfa/submit`: Endpoint for user corrections.
    * `/v1/nfa/verify`: Policy engine hook for token alignment.
* **Data Storage/State:**
    * Correction History (Append-only).

## 5. Alternatives Considered
* **Context Injection:** Rejected; agents can "forget" or ignore injected context.
* **Hard Stop/Restart:** Rejected; too disruptive to the reasoning workflow.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevents autonomous loops from bypassing human directives.
* **Observability:** Metrics on "Correction Rejection" rates by agents.

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
