# Design Doc: Agentic Security Posture Management (ASPM) Hub
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
With the resurgence of autonomous CLI agents like Claude Code and OpenClaw, enterprises are facing a "Non-Human Visibility Crisis." 48.9% of organizations are blind to agent-generated machine-to-machine (M2M) traffic. MCP Any needs a centralized hub to monitor, score, and attest to the security posture of all connected agents and tools to ensure compliance and mitigate supply-chain risks.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide real-time security posture scoring for all connected agents and MCP tools.
    *   Maintain a hardware-attested registry of "Known Good" agent configurations.
    *   Integrate with CLAW-10 compliance mapping for automated reporting.
*   **Non-Goals:**
    *   Acting as a full SIEM (Security Information and Event Management) replacement.
    *   Enforcing model-level alignment (handled by the LLM providers).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Verify that a newly deployed OpenClaw swarm complies with the organization's "Zero-Trust Retrieval" policy before granting access to production databases.
*   **The Happy Path (Tasks):**
    1.  The architect opens the MCP Any **ASPM Dashboard**.
    2.  The Hub automatically discovers the OpenClaw instances and performs a "Security Handshake."
    3.  The Hub scores the swarm based on its attestation signals (TPM status, Inode-pinning enablement, and provenance).
    4.  The architect reviews the "ASPM Scorecard" and sees that the swarm is compliant.
    5.  The Hub issues a hardware-bound "Posture Token" that allows the swarm to invoke the production database MCP tool.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Autonomous Agent] -->|Attestation Signals| ASPM_Hub[ASPM Hub]
        Tool[MCP Tool] -->|Behavioral Manifest| ASPM_Hub
        ASPM_Hub -->|Scorecard| Dashboard[Security Dashboard]
        ASPM_Hub -->|Posture Token| Policy_Engine[Zero-Trust Policy Engine]
    ```
*   **APIs / Interfaces:**
    *   `GET /api/v1/aspm/scorecard`: Retrieve security scores for all active sessions.
    *   `POST /api/v1/aspm/attest`: Submit hardware signals for posture verification.
*   **Data Storage/State:**
    *   Persistent SQLite store for historical posture trends and compliance logs.

## 5. Alternatives Considered
*   **Legacy WAFs:** Rejected because they lack "Semantic Awareness" of agent reasoning and intent.
*   **Manual Auditing:** Rejected due to the machine-speed nature of autonomous agent interactions.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Hub itself must be hardware-attested (TPM/SEP) to prevent "Posture Spoofing."
*   **Observability:** Integrated with the **M2M Visibility Engine** for real-time trace-linked auditing.

## 7. Evolutionary Changelog
*   **2026-04-11:** Initial Document Creation.
