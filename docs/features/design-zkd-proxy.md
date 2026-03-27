# Design Doc: Zero-Knowledge Discovery (ZKD) Proxy
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
As agent swarms become more autonomous and deep, the discovery phase has emerged as a primary reconnaissance vector. Malicious subagents can probe unmasked MCP schemas to map host surfaces, identify sensitive environment variables, and find high-value targets (e.g., shell tools) before any execution-time security policies can be enforced.

The **ZKD Proxy** provides "Privacy-Preserving Discovery" by cryptographically masking agent capability cards. It ensures that tool schemas remain invisible to peers until a verified, mission-bound handshake is completed, neutralizing pre-flight reconnaissance.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographic masking for all discovery-phase capability cards.
    * Mandate hardware-bound handshakes for unmasking tool schemas.
    * Provide "Schema Mirroring" defense by verifying tool metadata provenance.
    * Support mission-root anchored discovery policies.
* **Non-Goals:**
    * Encrypting tool call payloads (handled by transport-layer security).
    * Managing individual agent identities (handled by SMI/FSI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a specialized subagent from "mapping" host shell tools during the discovery phase.
* **The Happy Path (Tasks):**
    1. Architect enables ZKD Proxy for the local MCP Any node.
    2. Subagent initiates a discovery request for "available tools."
    3. MCP Any returns masked capability cards (e.g., hash-only identifiers) without schema details.
    4. Subagent attempts to probe a masked card.
    5. MCP Any interdicts the probe and requests a hardware-attested mission token.
    6. Subagent provides a valid token linked to its parent mission.
    7. ZKD Proxy unmasks the specific schemas authorized for that mission branch.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent] -->|List Tools| Proxy[ZKD Proxy]
        Proxy -->|Mask Schemas| Registry[MCP Registry]
        Registry -->|Masked Cards| Proxy
        Proxy -->|Return Masked| Agent
        Agent -->|Handshake + Token| Proxy
        Proxy -->|Verify Token| TPM[Hardware Enclave]
        TPM -->|Verified| Proxy
        Proxy -->|Unmask Authorized| Agent
    ```
* **APIs / Interfaces:**
    * `GetMaskedCapabilities()`: Returns a list of capability identifiers without structural metadata.
    * `UnmaskCapability(id, mission_token)`: Requests unmasking for a specific capability card.
* **Data Storage/State:**
    * Masked schema cache stored in the Ephemeral Workspace Trust middleware.

## 5. Alternatives Considered
* **Allow-list Only Discovery**: Rejected because it breaks autonomous "Teammate" discovery patterns where agents need to find specialists for new tasks.
* **Execution-time Gating Only**: Rejected as it allows reconnaissance; attackers can "see" tools even if they can't call them, providing clues for exfiltration or lateral movement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZKD is the primary defense against CVE-2026-91042 (Schema Mirroring).
* **Observability:** Discovery-phase "Reconnaissance Probes" (failed unmasking attempts) are logged as P0 security events.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
