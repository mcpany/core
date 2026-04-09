# Design Doc: Behavioral Redaction Gateway
**Status:** Draft
**Created:** 2026-07-24

## 1. Context and Scope
Traditional data masking (regex-based) is insufficient for AI agent data flows, where the sensitivity of information is context-dependent. A string that is safe for a "Log Auditor" might be highly sensitive for a "GitHub Triage" subagent. The Behavioral Redaction Gateway moves security from static patterns to context-aware sanitization based on the agent's active mission-root role and reasoning intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time, context-aware redaction of tool outputs and subagent coordination messages.
    * Dynamically adjust masking rules based on the "Current Role" verified by the mission-root.
    * Support "Zero-Trust Data Flows" where agents only see the minimum necessary state (Differential Privacy).
    * Neutralize "Context-Dump" exfiltration (CVE-2026-39102).
* **Non-Goals:**
    * Encrypting data at rest (handled by the storage layer).
    * Modifying the model's weights or internal reasoning logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Code Formatter" subagent from seeing private API keys in a config file, while allowing a "Security Auditor" subagent to verify them.
* **The Happy Path (Tasks):**
    1. Mission Root delegates "Format Code" to Agent A and "Audit Security" to Agent B.
    2. Agent A requests the content of `.env`.
    3. Redaction Gateway identifies Agent A's role and redacts all values, providing only keys.
    4. Agent B requests the same file.
    5. Redaction Gateway identifies Agent B's high-trust role and provides the raw content.
    6. Both agents receive data tailored to their intent without global exposure.

## 4. Design & Architecture
* **System Flow:**
    `Tool Output` -> `Context Extractor` -> `Role-Based Masking Engine` -> `Sanitized Result` -> `Agent`
* **APIs / Interfaces:**
    * `RedactionEngine`: `Sanitize(data []byte, role Role, intent Intent) ([]byte, error)`
* **Data Storage/State:**
    * Redaction policies stored in Rego/CEL.

## 5. Alternatives Considered
* **Static PII Scrubbers**: Rejected as they lack the flexibility to handle context-dependent sensitivity.
* **Manual HITL for Every Field**: Rejected due to prohibitive latency in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Default-masking for all unknown or low-trust roles.
* **Observability**: "Redaction Heatmap" shows which data fragments are being masked and why.

## 7. Evolutionary Changelog
* **2026-07-24:** Initial Document Creation.
