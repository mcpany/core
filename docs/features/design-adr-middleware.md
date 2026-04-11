# Design Doc: Agentic Detection and Response (ADR) Middleware
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
The rise of "Hivenet" swarm attacks and the exploitation of CLI agent local loopback ports (CVE-2026-25253) demand a response mechanism that operates at machine speed. ADR Middleware provides the "Immune System" for MCP Any, detecting anomalous reasoning patterns and coordination signals to autonomously interdict malicious agent activity.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Perform sub-millisecond, cross-agent behavioral analysis to detect coordinated attacks.
    *   Provide autonomous quarantine and capability revocation for compromised sessions.
    *   Support "Reasoning Noise-Injection" to disrupt side-channel probing attempts.
*   **Non-Goals:**
    *   Replacing human-in-the-loop (HITL) for high-stakes business approvals.
    *   Blocking legitimate high-entropy reasoning (e.g., complex code generation).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Protect the local environment from an OpenClaw specialist agent that has been "ClawJacked" and is attempting to exfiltrate session tokens.
*   **The Happy Path (Tasks):**
    1.  The compromised agent attempts to probe the local management interface.
    2.  The **ADR Middleware** detects the anomalous "Low-and-Slow" probing sequence.
    3.  ADR triggers an **Autonomous Quarantine** signal.
    4.  MCP Any immediately revokes the agent's "Session Scopes" and locks its mailbox.
    5.  The incident is logged in the **ADR Alerts Dashboard** for post-mortem analysis.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>ADR_Middleware: Coordination Fragment
        ADR_Middleware->>Behavior_Analyzer: Semantic Entropy Check
        Behavior_Analyzer-->>ADR_Middleware: Anomaly Score (Critical)
        ADR_Middleware->>Scope_Manager: Revoke Capability
        ADR_Middleware->>Alert_Hub: Log Interdiction
    ```
*   **APIs / Interfaces:**
    *   `POST /api/v1/adr/interdict`: Manually trigger a session quarantine.
    *   `GET /api/v1/adr/alerts`: Fetch real-time interdiction logs.
*   **Data Storage/State:**
    *   In-memory "Hot Store" for real-time behavioral baselines and entropy counters.

## 5. Alternatives Considered
*   **Static Rate Limiting:** Rejected because modern "Agentic Social Engineering" attacks can stay under standard thresholds.
*   **External SIEM Integration:** Rejected due to latency; interdiction must occur in the sub-millisecond range to be effective.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** ADR triggers must be non-repudiable and linked to hardware-attested mission roots.
*   **Observability:** Visualized via real-time "Swarm Anomaly" heatmaps in the UI.

## 7. Evolutionary Changelog
*   **2026-04-11:** Initial Document Creation.
