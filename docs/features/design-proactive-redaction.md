# Design Doc: Proactive Secret Redaction Middleware

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
The recent OpenClaw information disclosure vulnerability (CVE-2026-27003) showed that agents often accidentally leak sensitive information (like Telegram bot tokens, API keys, or PII) into plaintext logs or the LLM's context. Standard logging redaction is often "too late" as the secret has already entered the agent's memory or session state. MCP Any needs a proactive barrier that intercepts and redacts sensitive data at the middleware level, before it is processed by any other component.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all tool inputs and outputs.
    *   Apply regex-based and entropy-based pattern matching to identify secrets (API keys, tokens, PII).
    *   Redact identified secrets in-flight with a standardized placeholder (e.g., `[REDACTED_SECRET]`).
    *   Support configurable redaction rules per service or globally.
    *   Ensure redaction happens *before* logs are written and *before* data is returned to the calling Agent.
*   **Non-Goals:**
    *   Replacing dedicated secret management (like Vault).
    *   Encrypted storage of intercepted secrets (the goal is to remove them, not store them).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Agent Developer.
*   **Primary Goal:** Prevent a tool that fetches "System Info" from accidentally leaking the `DATABASE_URL` contained in an environment variable string it returned.
*   **The Happy Path (Tasks):**
    1.  The Redaction Middleware is enabled by default.
    2.  A tool returns a large JSON object containing a raw error message with an embedded API key.
    3.  The Middleware scans the JSON string, matches the API key pattern.
    4.  The Middleware replaces the key with `[REDACTED_API_KEY]`.
    5.  The calling Agent receives the "clean" JSON.
    6.  The system logs show only the redacted version.

## 4. Design & Architecture
*   **System Flow:**
    - **Intercept**: The middleware sits at the edge of the tool execution pipeline.
    - **Scan Engine**: A high-performance scanning engine uses a library of predefined patterns (e.g., AWS keys, Stripe keys, JWTs) and custom user-defined regex.
    - **Transformation**: Recursive walk of JSON structures to redact values while preserving structure.
*   **APIs / Interfaces:**
    - Configuration schema for `redaction_rules`.
    - Hook in the `ToolExecutor` pipeline.
*   **Data Storage/State:** Stateless transformation.

## 5. Alternatives Considered
*   **Post-hoc Log Redaction**: Scanning logs after they are written. *Rejected* because it doesn't protect the LLM context or session state.
*   **Agent-side Redaction**: Relying on the Agent (LLM) to redact its own output. *Rejected* as it is unreliable and vulnerable to prompt injection.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Essential for maintaining a Zero Trust perimeter where even trusted tools are monitored for accidental data leakage.
*   **Performance:** Scanning large tool outputs can add latency. Need to use optimized regex engines (e.g., RE2) and potentially limit scan depth/size.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
