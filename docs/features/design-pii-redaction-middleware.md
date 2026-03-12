# Design Doc: PII Redaction Middleware (Privacy Shield)
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
With the increasing use of AI agents in enterprise environments, a major concern is the accidental exposure of sensitive data—Personally Identifiable Information (PII), API keys, and internal secrets—to third-party LLM providers. The Meta/OpenClaw incident has accelerated the need for a "Privacy-First" approach. MCP Any needs a high-performance, mandatory middleware layer that scans tool outputs and redacts sensitive information before it leaves the local execution environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect and redact PII (emails, phone numbers, names) from tool outputs.
    * Identify and mask secrets (API keys, passwords, tokens) using pattern matching and entropy analysis.
    * Provide a configurable "Redaction Policy" (e.g., Allow, Mask, or Redact).
    * Maintain a secure audit log of all redactions for user verification.
* **Non-Goals:**
    * Redacting data within the LLM's primary prompt (this middleware focuses on *tool results*).
    * Guaranteed 100% detection of all possible sensitive data (it is a best-effort, heuristic-based shield).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Administrator.
* **Primary Goal:** Ensure that an agent querying a local database doesn't send customer email addresses to a cloud-based LLM.
* **The Happy Path (Tasks):**
    1. Agent executes `db_query` tool to fetch customer records.
    2. The tool returns a JSON array containing `name`, `email`, and `purchase_history`.
    3. The PII Redaction Middleware intercepts the output.
    4. Based on the `Enterprise-Strict` policy, it replaces all email addresses with `[REDACTED_EMAIL]`.
    5. The sanitized JSON is delivered to the Agent's context.
    6. The administrator can view the redacted data in the `Privacy Attestation Dashboard`.

## 4. Design & Architecture
* **System Flow:**
    `Tool Execution` -> `PII Redaction Middleware` -> `LLM / Agent Context`
    1. **Scanning**: Uses a combination of Regex, Presidio (or similar), and custom entropy-based secret detectors.
    2. **Transformation**: Modifies the tool output payload based on the matched entity types.
    3. **Metadata Injection**: Adds a header `X-MCP-Redacted: true` to inform the agent that data has been sanitized.
* **APIs / Interfaces:**
    * `Middleware.Process(payload)`: The core entry point for all tool results.
    * `GET /v1/privacy/audit`: Retrieve history of redactions for the current session.
* **Data Storage/State:**
    * `redaction_policies.yaml`: Configuration for detection rules and sensitivity levels.
    * `privacy_audit.db`: Temporary storage for redaction metadata (type, key, hash of original value).

## 5. Alternatives Considered
* **Agent-side Redaction**: Rejected because it relies on the agent being "well-behaved." Security must be enforced at the infrastructure level.
* **LLM-side Redaction (System Prompts)**: Unreliable; LLMs can be bypassed via prompt injection to reveal original data.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The middleware itself must run in a secure memory space. Redacted data is never stored in plain text.
* **Observability**: Real-time metrics on redaction frequency and latency are exported to the `Resource Telemetry Middleware`.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
