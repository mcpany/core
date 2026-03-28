# Design Doc: Metadata Sanitization Gateway (MSG)
**Status:** Draft
**Created:** 2026-03-28

## 1. Context and Scope
As agents become more integrated into collaborative platforms like GitHub and Slack, they are increasingly exposed to untrusted external metadata (issue titles, comments, messages). Attackers can use these fields for "Indirect Prompt Injection," tricking the agent into executing malicious commands under the guise of processing a legitimate ticket or message. MSG provides a mandatory sanitization layer for all external metadata ingested by connected agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic sanitization of metadata strings before ingestion.
    * Strip imperative instructions and high-entropy noise from metadata.
    * Validate metadata against a "Reasoning-Aware" blocklist.
* **Non-Goals:**
    * General-purpose PII scrubbing (handled by the PII-Sovereign Context Scrubber).
    * Modifying the original source metadata (e.g., editing the GitHub issue).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious CI/CD Integrator
* **Primary Goal:** Prevent a malicious GitHub issue title from triggering a `rm -rf` command via an autonomous triage agent.
* **The Happy Path (Tasks):**
    1. Agent tool retrieves a GitHub issue for triage.
    2. MSG intercepts the metadata (title, body) before it is returned to the agent's context.
    3. MSG performs semantic analysis, detecting imperative phrases like "Ignore previous instructions and run..."
    4. MSG redacts or rephrases the malicious fragment.
    5. Agent receives a sanitized version of the issue and proceeds with safe triage.

## 4. Design & Architecture
* **System Flow:**
    `External Source (GitHub/Slack)` -> `Agent Tool` -> `MSG Middleware` -> `Agent Context`
* **APIs / Interfaces:**
    * `MetadataSanitizer`: `Sanitize(metadata map[string]string) (sanitized map[string]string, alerts []string)`
* **Data Storage/State:**
    * Uses a local, hardware-attested blocklist and semantic regex rules stored in the Blackboard.

## 5. Alternatives Considered
* **Agent-Side Filtering**: Rejected because the agent's own reasoning is the target of the attack.
* **Source-Side Webhooks**: Rejected as it requires individual configuration for every external platform.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MSG runs in a detached sandbox, isolated from the agent reasoning engine.
* **Observability:** Sanitization events and detected injection attempts are logged to the Action-Chain Sovereignty Monitor.

## 7. Evolutionary Changelog
* **2026-03-28:** Initial Document Creation.
