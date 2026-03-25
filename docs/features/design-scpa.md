# Design Doc: Supply Chain Provenance Attestor (SCPA)
**Status:** Draft
**Created:** 2026-06-05

## 1. Context and Scope
Recent "Clinejection" and "FuncPoison" exploits have demonstrated that the agentic supply chain—specifically tool definitions and structural metadata—is the new primary attack vector. Attackers are injecting malicious instructions into tool descriptions that models treat as high-trust system guidance.

The SCPA is required to ensure that every tool and capability exposed via the Universal Agent Bus has a verifiable, hardware-attested provenance chain, neutralizing metadata-based reasoning hijacking.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement TPM-bound signing for all MCP tool definitions and schemas.
    * Provide real-time attestation verification during the tool discovery phase.
    * Support federated reputation quorums for third-party skill registries.
* **Non-Goals:**
    * Sandbox the execution of the tools itself (handled by Ephemeral Discovery Sandbox).
    * Perform semantic analysis of the tool's code (handled by Structural Metadata Sanitizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Architect
* **Primary Goal:** Ensure only verified, corporate-approved tools are visible to internal agent swarms.
* **The Happy Path (Tasks):**
    1. Administrator registers a new MCP server with the SCPA.
    2. SCPA generates a unique hardware-bound signature for the server's tool manifest.
    3. An agent requests tool discovery.
    4. MCP Any filters the discovery bus, only revealing tools with valid SCPA signatures.

## 4. Design & Architecture
* **System Flow:**
    [Registry] -> (Signature Request) -> [SCPA (TPM-bound)] -> (Signed Manifest) -> [Discovery Bus]
* **APIs / Interfaces:**
    * `POST /v1/attest/tool`: Submits tool metadata for signing.
    * `GET /v1/verify/tool/{id}`: Verifies the provenance signature of a capability.
* **Data Storage/State:**
    Signatures are stored in a versioned, immutable audit log within the local TPM-bound vault.

## 5. Alternatives Considered
* **Manual Allow-listing:** Rejected due to the scale of modern swarms and the high MTTC (Mean Time To Change) for manual reviews.
* **Purely Cloud-based Attestation:** Rejected to maintain "Local Sovereignty" and ensure offline resilience.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SCPA uses a "Deny by Default" posture for any unsigned capability discovery request.
* **Observability:** Every attestation event and verification failure is logged to the Mesh-Resident Audit Registry.

## 7. Evolutionary Changelog
* **2026-06-05:** Initial Document Creation.
