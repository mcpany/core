# Design Doc: Federated Identity Attestation
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As AI agents move from single-machine operations to federated swarms that span multiple organizations and network boundaries, the risk of "Agent Impersonation" and "Tool Hijacking" increases. MCP Any currently lacks a standardized way to cryptographically verify the identity of a remote MCP server or a connecting agent in a federated mesh.

This feature introduces a "Federated Identity Attestation" layer that uses cryptographic signatures (e.g., based on DID, SPIFFE, or X.509) to ensure that every participant in the MCP fabric is verified before tools are discovered or executed.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a standard interface for identity providers (IdPs) to attest to agent/server identity.
    *   Implement cryptographic verification of MCP server origins during discovery.
    *   Enable cross-boundary trust through federated trust bundles.
    *   Quarantine "Shadow" (unverified) tools by default.
*   **Non-Goals:**
    *   Implementing a new PKI (Public Key Infrastructure). We will leverage existing standards.
    *   Managing user-level authentication (this is for service-to-service/agent-to-agent identity).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Securely allow an external "Specialized Audit Agent" from a partner firm to access specific internal "Audit Tools" via MCP Any without compromising the entire infrastructure.
*   **The Happy Path (Tasks):**
    1.  The partner agent presents a signed Attestation Token during the MCP `initialize` handshake.
    2.  MCP Any verifies the token against the partner's public trust bundle.
    3.  MCP Any resolves the agent's identity to a specific "Partner Audit" profile.
    4.  The Policy Engine allows discovery of ONLY the audit-scoped tools.
    5.  The agent executes a tool; MCP Any logs the execution with the verified identity of the caller.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant Gateway as MCP Any Gateway
        participant TrustStore as Federated Trust Store
        participant Policy as Policy Engine

        Agent->>Gateway: initialize (with Attestation Token)
        Gateway->>TrustStore: Verify Signature & Expiry
        TrustStore-->>Gateway: Identity Verified (Firm A: Auditor)
        Gateway->>Policy: Get Permissions for "Firm A: Auditor"
        Policy-->>Gateway: Allow [read_logs, audit_db]
        Gateway-->>Agent: initialize response (Scoped Tools)
    ```
*   **APIs / Interfaces:**
    *   New `AttestationProvider` interface in `pkg/auth`.
    *   Updated `Handshake` middleware to intercept and verify attestation headers.
*   **Data Storage/State:**
    *   Trust bundles stored in the Service Registry or a secure KV store.
    *   Verified identities cached in the session state.

## 5. Alternatives Considered
*   **Static API Keys:** Rejected because they are easily leaked and do not provide provenance/identity of the *actual* caller in a complex swarm.
*   **mTLS (Mutual TLS):** A strong alternative, but often difficult to manage across different cloud providers and dynamic agent environments. We will support mTLS as one of the attestation providers.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** If attestation fails, the connection is dropped immediately. No "guest" access by default in federated mode.
*   **Observability:** All attestation successes and failures are logged to the Audit Trail with detailed metadata (Issuer, Subject, Fingerprint).

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
