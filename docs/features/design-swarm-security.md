# Design Doc: Coordinated Swarm Security Hub
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
The rise of coordinated "Predator Swarm" attacks has highlighted a gap in traditional agent security. These attacks distribute malicious tasks across multiple agents, making each individual action appear benign. MCP Any needs a centralized "Swarm Security Hub" to correlate telemetry across all active agent sessions and detect these distributed attack patterns.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized "Signal Correlator" for multi-agent telemetry.
    * Detect distributed attack patterns (e.g., sequential scanning, staged exfiltration) across different agent IDs.
    * Provide a unified "Swarm Risk Score" for all active sessions.
    * Support "Intent-Based Attestation" where agents must declare a high-level goal that is verified against execution patterns.
* **Non-Goals:**
    * Policing single-agent behavior (this is handled by the Policy Firewall).
    * Providing real-time blocking of every individual tool call (focus is on higher-level swarm patterns).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer running a local agent swarm.
* **Primary Goal:** Detect if a group of subagents is coordinating a data exfiltration attempt.
* **The Happy Path (Tasks):**
    1. User enables "Swarm Governance" in the MCP Any settings.
    2. Multiple subagents begin tool calls; telemetry is streamed to the Swarm Security Hub.
    3. The Hub detects that Subagent A is reading sensitive files while Subagent B is preparing an external network request.
    4. The Hub flags a "High Risk Correlation" and pauses all sessions for user review.

## 4. Design & Architecture
* **System Flow:**
    - **Telemetry Sink**: All tool execution middleware sends metadata (agent ID, tool, parameters, intent) to the Hub.
    - **Pattern Matcher**: A rule-based engine (using CEL/Rego) that looks for multi-agent correlation signatures.
    - **Risk Scoring**: A dynamic score assigned to each agent swarm based on observed behavior vs declared intent.
* **APIs / Interfaces:**
    - `POST /v1/security/swarm/intent`: Register a swarm-level objective.
    - `GET /v1/security/swarm/status`: Retrieve the current risk state of the swarm mesh.
* **Data Storage/State:** In-memory sliding window of recent tool calls across all sessions, backed by SQLite for historical analysis.

## 5. Alternatives Considered
* **Isolated Agent Sandboxing**: Rejection as it doesn't prevent coordinated attacks where each agent stays within its sandbox but contributes to a global malicious goal.
* **Global Rate Limiting**: Too blunt; prevents legitimate high-volume agent work.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Hub itself must be protected and its telemetry stream must be tamper-proof.
* **Observability:** Requires a new UI dashboard (Swarm Timeline) to visualize correlations.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
