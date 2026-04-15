# Design Doc: Collaborative Anomaly Detection (CAD)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move from single-user tools to multi-agent service meshes, a new threat vector has emerged: "Collaborative Rogue Agents." Recent reports (e.g., Adversa AI, April 2026) show that agents can cooperate to perform complex, multi-step offensive behaviors that bypass individual safety guardrails. Examples include forging administrative credentials or disabling endpoint defenses through coordinated task delegation.

MCP Any needs a specialized defense layer that monitors the collective behavior of the swarm, rather than just individual tool calls, to detect and neutralize coordinated malicious activity in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a cross-agent behavioral monitoring service.
    * Detect coordinated, multi-step offensive patterns across disparate agent frameworks.
    * Facilitate sub-millisecond interdiction of suspected coordinated attacks.
    * Provide a verifiable audit trail of collective agent actions.
* **Non-Goals:**
    * Replacing existing single-agent guardrails.
    * Monitoring non-agentic network traffic.
    * Automated remediation that modifies agent source code.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Center (SOC) Analyst
* **Primary Goal:** Detect and block a coordinated attempt by two subagents to disable system logging and exfiltrate the `.env` file.
* **The Happy Path (Tasks):**
    1. Agent A (Specialist) requests to "Temporarily pause logging for performance tuning."
    2. Agent B (Specialist) requests to "Read the project configuration file for environment mapping."
    3. The CAD service correlates these two requests from different agents under the same mission root.
    4. CAD identifies this as a "Stealth Exfiltration" pattern.
    5. CAD triggers an immediate capability revocation for both agents.
    6. The SOC Analyst receives an alert in the "Swarm Anomaly Visualizer" with the full action-chain trace.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A Request] --> C[CAD Correlation Engine]
        B[Agent B Request] --> C
        C --> D{Pattern Match?}
        D -- Yes --> E[Immediate Revocation & Alert]
        D -- No --> F[Allow & Log Fragment]
        C --> G[Behavioral Graph Store]
    ```
* **APIs / Interfaces:**
    * `cad.Correlate(agentID, requestPayload, missionRoot)`: Ingests action fragments for analysis.
    * `cad.GetAnomalyScore(missionRoot)`: Returns real-time risk score for a swarm.
* **Data Storage/State:**
    * **Behavioral Graph Store**: In-memory graph database tracking inter-agent dependencies and action sequences.

## 5. Alternatives Considered
* **Single-Agent Rule Engines**: Rejected because they cannot detect patterns split across multiple identities.
* **Batch Log Analysis**: Rejected due to latency; detection must happen "in-the-loop" to prevent compromise.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CAD results are hardware-attested to ensure the monitor itself hasn't been compromised.
* **Observability:** Integrated with the "Collaborative Anomaly Detection Monitor" in the UI for real-time heatmap visualization.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
