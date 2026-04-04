# Design Doc: Metadata Provenance Engine
**Status:** Draft
**Created:** 2026-04-04

## 1. Context and Scope
Recent vulnerabilities (CVE-2026-42001) demonstrate that "Metadata Context Poisoning" is a critical risk where unauthenticated tool schemas (JSON-RPC definitions) are weaponized to inject malicious instructions into the agent's reasoning loop. The Metadata Provenance Engine provides a "Structural Attestation" layer to ensure every tool definition is cryptographically bound to a verified source.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce mandatory cryptographic signing for all MCP tool schemas (descriptions, arguments, examples).
    * Maintain a "Verified Metadata Lineage" (VML) for every discovered tool.
    * Automatically quarantine any tool whose structural metadata has been modified without a valid re-signature.
* **Non-Goals:**
    * Validating the *logic* inside the tool implementation (covered by other safety gates).
    * Managing the execution of the tool itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Prevent "Ghost Instructions" from entering the LLM context via high-trust tool definitions.
* **The Happy Path (Tasks):**
    1. A new MCP server is added to the gateway.
    2. The Metadata Provenance Engine intercepts the `tools/list` response.
    3. The Engine verifies the cryptographic signature of the tool schemas against the developer's public key.
    4. Upon successful verification, the tool is marked as "Attested" and exposed to the LLM.
    5. If a schema is modified (e.g., via a poisoned registry update), the Engine detects the signature mismatch and revokes the tool.

## 4. Design & Architecture
* **System Flow:**
    `MCP Server` -> `Schema Interception` -> `VML Signature Validator` -> `Attested Registry` -> `LLM`
* **APIs / Interfaces:**
    * `mcpserver.ToolsMiddleware`: Injects provenance checks into the tool discovery lifecycle.
    * `Registry API`: Managed by the PNTD Provider to track structural signatures.
* **Data Storage/State:**
    * `vml_registry.db`: SQLite database mapping tool IDs to their last-verified SHA-256 signatures and public keys.

## 5. Alternatives Considered
* **Runtime Sanitization Only**: Rejected because LLMs can be tricked even by subtle changes in tool "examples" that are difficult to scan semantically in real-time.
* **Hardcoded Allow-lists**: Too rigid for dynamic agent ecosystems.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Structural integrity is a prerequisite for discovery. Unsigned schemas are treated as "high-risk" and quarantined.
* **Observability:** Visual indicator in the UI (Metadata Provenance Viewer) showing the signing status of every available tool.

## 7. Evolutionary Changelog
* **2026-04-04:** Initial Document Creation.
