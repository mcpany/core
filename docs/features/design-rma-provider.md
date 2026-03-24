# Design Doc: Recursive Mission Attestation (RMA) Provider
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
As agent swarms grow in depth (parent -> child -> grandchild), the original user intent (Mission Root) often becomes diluted or "spliced" by intermediate subagents. Recursive Mission Attestation (RMA) provides a cryptographic "Chain of Custody" for intent, ensuring every sub-task is explicitly authorized by the root mission.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Cryptographically link every sub-task to the mission-root intent.
    *   Provide hardware-attested "Mission Receipts" for every delegation.
    *   Neutralize "Intent Splicing" and unauthorized boundary expansion.
*   **Non-Goals:**
    *   Enforce the *content* of the reasoning (that is for the PBR Adapter).
    *   Manage transport-layer security (handled by TLSB).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Architect
*   **Primary Goal:** Verify that a tool call made by a 4th-hop subagent was actually authorized by the original user request.
*   **The Happy Path (Tasks):**
    1.  User initiates mission; MCP Any issues a Hardware-Attested Mission Root Token.
    2.  Agent A delegates to Agent B; MCP Any issues a child "Mission Receipt" bound to the root.
    3.  Agent B calls a high-trust tool.
    4.  MCP Any validates the tool call by recursively verifying the Mission Receipt chain back to the Root Token.

## 4. Design & Architecture
*   **System Flow:**
    `Mission Root (TPM Signed) -> Child Receipt (A) -> Child Receipt (B) -> Tool Call Validator`
*   **APIs / Interfaces:**
    *   `POST /v1/mission/receipt`: Issue a new child receipt.
    *   `GET /v1/mission/verify`: Validate a receipt chain.
*   **Data Storage/State:**
    Receipts are stored as cryptographically signed JWS (JSON Web Signatures) in the Blackboard's metadata shard.

## 5. Alternatives Considered
*   **Flat Capability Tokens:** Rejected because they don't provide lineage; a stolen token could be used outside its intended mission branch.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Receipts are hardware-bound and mission-specific.
*   **Observability:** All receipt issuance and verification events are logged to the Audit Sink.

## 7. Evolutionary Changelog
*   **2026-06-07:** Initial Document Creation.
