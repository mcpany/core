// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

# Design Doc: Negative Feedback Attestation (NFA) Middleware

**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
Autonomous agent loops often suffer from "Reasoning Stubbornness," where an agent continues to repeat a failing path despite explicit user correction. This leads to wasted tokens and user frustration. Gemini CLI v1.5 has introduced Negative Feedback Attestation (NFA) to solve this.

MCP Any needs to implement NFA Middleware to ensure that user corrective feedback is cryptographically bound to the agent's session state. This makes it impossible for the agent to "ignore" feedback in subsequent reasoning steps without triggering a policy violation.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept user-provided "Negative Feedback" or "Corrections."
    * Cryptographically bind feedback to the current session's "Intent Chain."
    * Enforce that subsequent tool calls or reasoning steps acknowledge the feedback.
    * Provide a "Feedback Violation" signal if the agent diverges from the correction.
* **Non-Goals:**
    * Automatically rewriting the agent's prompt (handled by the LLM).
    * Restricting "Positive Feedback" (only negative/corrective feedback is attested).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an autonomous refinement loop.
* **Primary Goal:** Correct an agent's persistent mistake (e.g., using the wrong API version) and ensure it doesn't repeat it.
* **The Happy Path (Tasks):**
    1. The agent proposes a tool call with an incorrect parameter.
    2. The user provides negative feedback via the A2UI Gateway: "Stop using v1, use v2."
    3. MCP Any NFA Middleware signs the feedback and injects it into the mission root.
    4. The agent attempts to call the tool again with v1.
    5. The NFA Middleware detects the violation (mismatch with attested feedback).
    6. The tool call is blocked; the agent is forced to re-plan with the correction.

## 4. Design & Architecture
* **System Flow:**
    `User -> [A2UI Gateway] -> [NFA Middleware] -> [Signed Feedback Store] -> [Policy Engine]`
* **APIs / Interfaces:**
    * `POST /v1/feedback/negate`: Submit corrective feedback for a session.
    * `X-MCP-NFA-Status`: Header indicating if the current request complies with attested feedback.
* **Data Storage/State:**
    * Feedback fragments stored in the Shared KV Store (Blackboard) with session-bound isolation.
    * Cryptographic signatures managed by the local Identity Store.

## 5. Alternatives Considered
* **Prompt-Only Corrections**: Rejected because agents often "forget" or deprioritize instructions in long context windows. Attestation provides a hard infrastructure boundary.
* **Total Session Reset**: Rejected as it loses all progress and is inefficient.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Feedback must be signed by the user or an authorized supervisor agent.
* **Observability**: Feedback alignment scores visualized in the NFA Feedback Compliance Viewer.

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
