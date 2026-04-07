# Design Doc: Atomic Session-Revocation Enforcer (ASRE)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of CVE-2026-34503 in OpenClaw has highlighted a critical gap in agent security: session persistence beyond token revocation. Currently, even when a user refreshes their tokens or removes a device, active WebSocket and Named-Pipe sessions can remain open, allowing an attacker to continue executing tools via the gateway.

ASRE is a kernel-level enforcement layer for MCP Any that ensures session termination is atomic and universal across all transport channels. It bridges the gap between the Identity Provider (FSI) and the coordination bus, ensuring that a "Revoke" signal results in immediate closure of all physical communication paths.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-millisecond termination of active WebSocket and Named-Pipe sessions upon revocation.
    * Mandate hardware-attested heartbeats for session liveness.
    * Synchronize revocation signals across distributed MCP Any nodes in a mesh.
* **Non-Goals:**
    * Re-authenticating sessions (this is handled by FPIR).
    * Managing the underlying identity lifecycle (handled by NHI Governance).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Administrator
* **Primary Goal:** Immediately neutralize a suspected agent compromise by revoking all active sessions.
* **The Happy Path (Tasks):**
    1. Administrator identifies an anomaly and triggers "Emergency Revoke" via the FSI Provider.
    2. FSI Provider broadcasts a hardware-signed Revocation Token to the ASRE Hub.
    3. ASRE identifies all active session IDs (WebSockets, Named Pipes) associated with the compromised identity.
    4. ASRE forcefully closes the kernel-level file descriptors associated with those transports.
    5. Teammates attempt to use the coordination bus and receive an "Attestation Failed: Revoked" error.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        FSI->>ASRE: Broadcast RevocationToken (signed)
        ASRE->>CoordBus: Lookup SessionIDs
        ASRE->>Kernel: close(FDs)
        CoordBus-->>Agent: Session Closed
    ```
* **APIs / Interfaces:**
    * `POST /v1/security/revoke`: Internal endpoint for ingestion of revocation signals.
    * `GET /v1/security/sessions`: Monitor active session/FD mappings.
* **Data Storage/State:**
    * ASRE maintains an in-memory map of `IdentityID -> [SessionID, FileDescriptor]`, backed by a hardware-locked SQLite fragment for persistence across crashes.

## 5. Alternatives Considered
* **Application-Level Polling:** Rejected due to the "Revocation Lag" pain point (minutes vs. milliseconds).
* **TLS Session Ticket Expiration:** Rejected because it doesn't handle active, long-lived WebSockets or local Named Pipes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The revocation signal itself must be TPM-signed to prevent Denial-of-Service attacks by malicious subagents.
* **Observability:** ASRE logs all termination events with high-resolution timestamps to the Local Security Audit Log.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
