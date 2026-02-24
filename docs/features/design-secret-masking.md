# Design Doc: Secret Masking Middleware

**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
The OWASP MCP Top 10 (2025) identifies "Token Mismanagement & Secret Exposure" as a critical risk. AI agents often process sensitive authentication tokens, API keys, and PII. These secrets can easily leak into protocol logs, debugging traces, or even be returned to the LLM in tool outputs. The Secret Masking Middleware provides a centralized, high-performance redaction layer to ensure that sensitive data never leaves the secure boundary of MCP Any.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect and redact standard secret patterns (API keys, Bearer tokens, etc.) in logs and traces.
    * Allow users to define custom regex patterns for domain-specific sensitive data.
    * Mask secrets in tool inputs/outputs before they are persisted or sent to an LLM.
    * Support "Zero-Knowledge" logging where only hashes of secrets are logged for debugging purposes.
* **Non-Goals:**
    * Implementing a secret manager (MCP Any should integrate with existing vaults).
    * Encryption of data at rest (handled by the storage layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer.
* **Primary Goal:** Prevent an agent from accidentally leaking a GitHub Personal Access Token into the centralized logging system.
* **The Happy Path (Tasks):**
    1. Engineer enables the `SecretMaskingMiddleware` in the MCP Any config.
    2. An agent calls a tool that requires a `GITHUB_TOKEN`.
    3. The middleware intercepts the tool call.
    4. When MCP Any logs the transaction to the database, the `GITHUB_TOKEN` value is replaced with `[REDACTED_GH_TOKEN]`.
    5. In the UI Traces view, the token is likewise masked.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: Middleware sits between the `ToolEngine` and the `Logger/Tracer` components.
    - **Detection**: Uses a multi-stage approach:
        - Known schema fields marked as `sensitive`.
        - Regex-based pattern matching for common formats.
        - High-entropy string detection.
    - **Transformation**: Replaces matching strings with redaction markers or truncated hashes.
* **APIs / Interfaces:**
    - `Mask(input string) string`: Core function for redacting strings.
    - `RegisterPattern(name string, regex string)`: API for adding custom patterns.
* **Data Storage/State:** Pattern configurations are stored in the main `config.yaml`.

## 5. Alternatives Considered
* **Client-side Redaction**: Redacting in the UI. *Rejected* because it doesn't protect the backend logs or the LLM context.
* **Manual Redaction in Tools**: Requiring tool authors to mask secrets. *Rejected* because it is error-prone and doesn't cover protocol-level leaks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The masking engine must itself be secure and not leak secrets through its own errors.
* **Observability:** Metrics should track how many secrets are redacted, providing a "Leak Prevention" dashboard.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation.
