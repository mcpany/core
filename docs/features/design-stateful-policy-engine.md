# Design Doc: Stateful Policy Engine (Guardrail API)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Traditional MCP policies are static, based on identity or fixed capabilities. As agents interact more deeply with users and sensitive data, we need "Stateful Policies" that adjust in real-time. For example, if an agent's "Safety Score" (calculated by an LLM or observer) drops due to suspicious requests, its permission to use destructive tools should be automatically revoked.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Integrate real-time "Safety Scores" into policy evaluation.
    *   Provide a "Guardrail API" for external observers to report agent behavior.
    *   Enable "Conditional Permissions" that depend on the current conversation's sentiment or safety status.
    *   Allow for "Automatic Downgrade" of permissions without manual intervention.
*   **Non-Goals:**
    *   Replacing the primary LLM's own safety filters.
    *   Providing a full-blown sentiment analysis engine (we ingest scores, we don't necessarily generate them).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin.
*   **Primary Goal:** Automatically block an agent from accessing internal databases if it starts attempting to exfiltrate data.
*   **The Happy Path (Tasks):**
    1.  An external Security Observer (monitoring the tool stream) detects a prompt injection attempt.
    2.  The Observer calls the MCP Any Guardrail API: `POST /v1/guardrails/report` with a `safety_score: 0.2`.
    3.  The Stateful Policy Engine receives the score and matches it against a rule: `if safety_score < 0.5 then revoke('db:*')`.
    4.  The agent's next attempt to call `db:query` is blocked by the Policy Firewall.
    5.  The incident is logged and the Admin is notified.

## 4. Design & Architecture
*   **System Flow:**
    *   Policy Engine maintains a "State Map" for each agent session.
    *   Guardrail API updates the state map.
    *   The Policy Firewall (Rego) is extended to query this State Map during evaluation.
*   **APIs / Interfaces:**
    *   `POST /v1/guardrails/report`: `{ session_id: string, safety_score: float, reasoning: string }`.
    *   Rego Extension: `input.state.safety_score`.
*   **Data Storage/State:**
    *   In-memory TTL cache for session state, backed by Shared KV Store for persistence.

## 5. Alternatives Considered
*   **Static Policies**: Rejected because they can't handle dynamic threats like prompt injection or social engineering.
*   **LLM-in-the-Middle**: Too slow and expensive for every tool call. A side-channel Guardrail API is more performant.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Guardrail API itself must be secured with high-privilege credentials.
*   **Observability:** Safety score trends are visualized in the Dynamic Policy Monitor.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
