# Design Doc: Headless Handoff Attestation (HHA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of "Remote Control" and "Dispatch" patterns in modern agent frameworks like Claude Code and OpenClaw, AI agents are increasingly decoupled from a single, persistent controller terminal. Users now initiate missions on one device (e.g., local CLI) and expect to monitor or steer them from another (e.g., mobile GUI or remote web dashboard).

MCP Any needs to solve the "Handoff Trust" problem. Transferring control of a running, high-privilege agent session between controllers without re-authenticating the entire mission root or exposing raw environment variables is a critical security frontier. HHA provides the cryptographic protocol to securely "hand over" the session baton.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable secure, hardware-attested transfer of agent control between disparate controllers.
    * Maintain "Mission Continuity" without restarting the agent ReAct loop.
    * Ensure that the new controller inherits only the necessary "Intent-Scoped" permissions.
    * Provide an audit trail for all session handovers.
* **Non-Goals:**
    * Multi-user concurrent control of a single agent session (current model is one active controller).
    * General-purpose remote desktop or terminal access.

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile Developer Steer-er
* **Primary Goal:** Take control of a coding agent mission initiated on a desktop CLI while commuting.
* **The Happy Path (Tasks):**
    1. User starts `mcpany dispatch` on their office workstation to run a long-running refactor.
    2. Workstation MCP Any instance generates an HHA "Handover Intent" token.
    3. User opens the MCP Any Mobile App and scans a secure QR code (or inputs an OOB code).
    4. Mobile app performs a hardware-bound attestation (FaceID/Secure Enclave) and sends the HHA request to the workstation.
    5. Workstation verifies the attestation and transfers the "Controller Baton" to the mobile device.
    6. Agent continues reasoning, now reporting traces and accepting steering from the mobile app.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant C1 as Original Controller (CLI)
        participant GW as MCP Any Gateway (Process)
        participant C2 as New Controller (Mobile)

        C1->>GW: Request Handoff (Intent: Steer)
        GW-->>C1: Issue HHA-Challenge (TPM-signed)
        C1-->>C2: Out-of-Band Transfer (QR/Code)
        C2->>GW: HHA-Response (Device Attestation + Challenge)
        GW->>GW: Verify Lineage & Hardware Root
        GW-->>C2: Grant Controller Baton
        GW-->>C1: Revoke Controller Baton
        GW->>C2: Resume Trace Streaming
    ```
* **APIs / Interfaces:**
    * `POST /v1/session/handoff/init`: Initiates a handoff request, returns a signed challenge.
    * `POST /v1/session/handoff/accept`: Consumes an attestation proof and completes the handover.
* **Data Storage/State:**
    * HHA tokens are stored in the memory-resident session state, bound by the `Universal Gateway Persistence` process.

## 5. Alternatives Considered
* **Persistent WebSockets with Long-lived JWTs:** Rejected because it doesn't provide hardware-bound proof of the *new* controller's identity and is vulnerable to token theft.
* **Re-authenticating via Mission-Root Token:** Rejected because it requires the user to manage high-privilege root tokens on every device, increasing the attack surface.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All handoffs require hardware-level attestation (TPM on desktops, Secure Enclave on mobile). Handover tokens are one-time-use and expire within 60 seconds.
* **Observability:** Handoff events are logged to the `Mesh-Resident Lineage Tracker` with metadata about the source and target controller hardware IDs.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
