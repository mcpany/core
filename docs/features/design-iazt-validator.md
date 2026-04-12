# Design Doc: Inter-Agent Zero-Trust (IAZT) Validator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The proliferation of horizontal agent swarms (e.g., Claude Code Agent Teams, OpenClaw specialist meshes) has exposed a critical "Lateral Infection" vulnerability. In most current implementations, once a researcher or specialist agent is compromised, it can silently infect peers or parent agents through unverified internal coordination channels.

The **Inter-Agent Zero-Trust (IAZT) Validator** is a core security component for MCP Any that mandates continuous authentication and role-based verification for every agent-to-agent (A2A) interaction. It treats internal swarm traffic as potentially hostile, requiring cryptographic proof of role-bound authority before any task is delegated or context is shared.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce cryptographically signed handshakes for all internal swarm coordination.
    *   Validate agent requests against a "Mission Role" policy (e.g., Researcher, Executor, Auditor).
    *   Detect and block lateral movement attempts from low-trust to high-trust agent sessions.
    *   Provide a standardized "Swarm Identity" token that persists across framework boundaries.
*   **Non-Goals:**
    *   Managing agent-to-LLM transport security (handled by the exfiltration gateway).
    *   Replacing individual agent sandboxing (e.g., gVisor, Docker).
    *   Defining the specific reasoning logic of the agents.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Prevent a compromised "Internet Researcher" agent from injecting malicious payloads into the "Database Admin" specialist agent.
*   **The Happy Path (Tasks):**
    1.  The Researcher agent completes its task and attempts to send a data fragment to the Database agent via the shared mailbox.
    2.  The IAZT Validator intercepts the request.
    3.  It verifies the Researcher's hardware-attested identity and checks if it is authorized to communicate with the Database agent for this specific mission phase.
    4.  The Validator performs a semantic check on the payload to ensure it doesn't contain re-engineered autonomous threats.
    5.  Upon successful validation, the request is cryptographically signed by IAZT and delivered to the Database agent's mailbox.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        A1[Researcher Agent] -->|Coordination Request| IAZT{IAZT Validator}
        IAZT -->|Check Role| Policy[Role Policy Store]
        IAZT -->|Verify Identity| TPM[Hardware Root]
        IAZT -->|Sanitize| APD[Adaptive Payload Detector]
        IAZT -->|Validated Request| A2[Database Agent]
    ```
*   **APIs / Interfaces:**
    *   `POST /iazt/handshake`: Establishes a verified role-bound session between teammates.
    *   `POST /iazt/verify`: Middleware endpoint for validating inter-agent messages.
    *   `Header: x-iazt-role-token`: Mandatory header for all mesh-bound traffic.
*   **Data Storage/State:**
    *   Role assignments and mission-bound trust tickets are stored in the secure session store (Redis or encrypted SQLite).
    *   Validation events are logged to the Action-Chain Sovereignty Monitor.

## 5. Alternatives Considered
*   **Network-Level Isolation (VLANs/Pipes)**: Rejected as insufficient because it cannot perform the semantic inspection required to stop context-window based infection.
*   **Static Manifests**: Rejected as they are too rigid for dynamic swarms; IAZT provides the necessary "continuous authentication" model.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All tokens are mission-scoped and hardware-bound. Replay attacks are neutralized via monotonic handshakes.
*   **Observability:** Integrated with the "Swarm Anomaly Visualizer" to show real-time lateral move alerts.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
