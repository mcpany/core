# Design Doc: Supply Chain Provenance Attestor (SCPA)
**Status:** Draft
**Created:** 2026-06-05

## 1. Context and Scope
The rapid expansion of the agentic ecosystem has led to a critical reliance on third-party tool registries and dynamic library updates. Recent security reports (Barracuda 2026) have identified over 43 agent framework components with backdoors introduced via supply chain compromise. The "Clinejection" attack pattern demonstrates how unverified tool definitions and library updates can be weaponized to exfiltrate credentials and PII.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate hardware-bound (TPM/Secure Enclave) signatures for all tool manifest updates.
    * Implement "Provenance-before-Discovery" validation for all A2A capability cards.
    * Provide a "Trust Strength" weight for the Active Negotiation Broker bidding process.
* **Non-Goals:**
    * Automated patching of compromised tools (SCPA focuses on detection and blocking).
    * Managing host-level binary signatures (SCPA focuses on the MCP tool definition layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an agent from discovering and using a "Shadow Tool" introduced via a compromised registry.
* **The Happy Path (Tasks):**
    1. Architect configures MCP Any to require "P0 Upstream Provenance."
    2. A new tool is discovered via the A2A mesh.
    3. SCPA intercept the discovery event and validates the Upstream Signature against the verified root.
    4. The signature check passes; the tool is added to the Discovery Bus.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: The SCPA middleware hooks into the PNTD Provider's discovery pipeline.
    - **Attestation**: It extracts the TPM-bound signature from the capability card and verifies it against the configured Upstream Registry Root.
    - **Scoring**: It calculates a "Provenance Score" (0-100) based on the signature strength and registry reputation.
* **APIs / Interfaces:**
    - POST /v1/provenance/validate: Core validation endpoint for capability cards.
    - Metadata: _mcp_provenance_token: string
* **Data Storage/State:** Local cache of verified registry public keys and revocation lists.

## 5. Alternatives Considered
* **Runtime Sandboxing Only**: Relying solely on Docker/WASM to contain tools. *Rejected* as this doesn't prevent "Logical Injection" or exfiltration of context-allowed data.
* **Manual Tool Allow-listing**: *Rejected* as it doesn't scale with the speed of autonomous agent meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SCPA implements "Provenance-before-Discovery." An agent cannot even "see" a tool in the discovery bus unless it carries a valid, hardware-attested provenance token.
* **Observability:** Every signature validation and provenance failure is logged to the Local Security Audit Dashboard.

## 7. Evolutionary Changelog
* **2026-06-05:** Initial Document Creation.
