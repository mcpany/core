# Design Doc: Hardware-Locked Environmental Handshake (HLEH)
**Status:** Draft
**Created:** 2026-06-17

## 1. Context and Scope
The emergence of **Identity Leakage via Process Environment (ILPE)** and the shift toward horizontal teammate coordination in heterogeneous swarms (Claude Code teammates vs. OpenClaw specialists) confirm that transport-layer security is insufficient. We must now protect the **environmental sovereignty** of the hardware-attested identity.

The **Hardware-Locked Environmental Handshake (HLEH)** is an authoritative gateway for all high-trust tool execution. It mandates hardware-bound proof of environmental purity and mission-root alignment before any capability is granted.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Mandate hardware-bound (TPM/SEP) proof of environmental purity.
    *   Protect mission-root identity from environmental exfiltration (ILPE).
    *   Synchronize environmental sovereignty across multi-cloud meshes.
    *   Provide sub-millisecond handshake latency for autonomous swarms.
*   **Non-Goals:**
    *   Replacing the entire OS process isolation.
    *   Managing non-security-critical environment variables.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Execute a high-privilege Python script on a remote specialist agent without leaking the mission-root identity token to the remote process environment.
*   **The Happy Path (Tasks):**
    1.  The agent requests execution of a privileged script via a remote Command Adapter.
    2.  The HLEH Gateway intercepts the request.
    3.  HLEH performs a hardware-locked (TPM) handshake with the target execution environment.
    4.  Target environment provides signed proof of environmental purity (ESE-compliant).
    5.  HLEH validates the proof and grants the session-bound capability.
    6.  Mission-root identity remains sovereign across the mesh boundary.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent] -->|Request Capability| HLEH[HLEH Gateway]
        HLEH -->|Challenge| TargetEnv[Target execution environment]
        TargetEnv -->|ESE-Signed Proof| HLEH
        HLEH -->|Validate & Grant| Agent
        Agent -->|Execute| TargetEnv
    ```
*   **APIs / Interfaces:**
    *   `ChallengeEnvironment(targetID string) (Nonce, error)`: Initiates a hardware-bound environmental challenge.
    *   `x-mcpany-hleh-proof`: New header for transporting real-time, hardware-attested environmental proofs.
*   **Data Storage/State:**
    *   Maintains a session-bound, hardware-locked manifest of all validated environment handshakes.

## 5. Alternatives Considered
*   **Transport Security Only (mTLS):** Rejected because mTLS cannot verify the state of the *environment* where the tool finally executes.
*   **Manual SSH Keys:** Rejected due to the machine-speed nature of agent swarms and the lack of automated environmental attestation.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** HLEH requires hardware-bound (TPM) signatures for all environmental proofs.
*   **Observability:** Integrated with the `Audit Log`, providing "Environmental Purity" scores.

## 7. Evolutionary Changelog
*   **2026-06-17:** Initial Document Creation.
