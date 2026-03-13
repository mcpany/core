# Design Doc: Proof-of-Intent (PoI) Validator
**Status:** Draft
**Created:** 2026-03-23

## 1. Context and Scope
With the rise of "Context-Mirroring" attacks (CVE-2026-34015), it's no longer enough to verify if an agent has permission to call a tool. We must verify that the tool call itself is a legitimate consequence of the agent's stated and signed "Intent." PoI Validator implements the UACO v1.7 standard to bind tool calls to cryptographically signed intent tokens, ensuring that agents stay within the logical boundaries of their assigned tasks.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate UACO v1.7 PoI headers on all tool calls.
    * Prevent "Context-Mirroring" where subagents are manipulated into leaking parent state via reflection tools.
    * Provide a cryptographic link between a high-level user goal and low-level tool execution.
* **Non-Goals:**
    * Perform full semantic analysis of the agent's reasoning (this is handled by the Policy Engine).
    * Replace the existing capability-based access control (RBAC/CBAC).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Admin
* **Primary Goal:** Ensure that a subagent performing "code review" cannot call a "file write" tool unless it's part of an explicit "apply fixes" intent signed by the parent agent.
* **The Happy Path (Tasks):**
    1. Parent agent establishes a session and signs an "Intent Token" for a specific task (e.g., "Review Code").
    2. Parent delegates a sub-task to a subagent, passing the signed Intent Token.
    3. Subagent attempts to call a tool (e.g., `read_file`).
    4. MCP Any PoI Validator intercepts the call and verifies the `X-UACO-PoI` header against the session's signed intent.
    5. The validator confirms the tool call is semantically and cryptographically linked to the "Review Code" intent.
    6. Tool execution proceeds.

## 4. Design & Architecture
* **System Flow:**
    `[UACO Client] -> [PoI Middleware] -> [Intent Verification] -> [Capability Check] -> [MCP Server]`
* **APIs / Interfaces:**
    * New `X-UACO-PoI` header containing a JWS (JSON Web Signature) of the intent, session context, and tool call hash.
    * `verifyIntent(token, toolCall)` internal API.
* **Data Storage/State:**
    * Ephemeral session-bound public keys for intent verification.
    * Intent-to-Tool mapping cache to reduce verification latency.

## 5. Alternatives Considered
* **Full Prompt Sanitization:** Rejected due to high latency and the "cat and mouse" nature of prompt injection defense.
* **Manual HITL for every subagent call:** Rejected as it breaks the autonomy and scalability of agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Use hardware-bound keys (TPM/Secure Enclave) where available for signing intents.
* **Observability:** Log PoI verification failures with detailed context (intent vs. tool call) for audit trails and RL feedback.

## 7. Evolutionary Changelog
* **2026-03-23:** Initial Document Creation.
