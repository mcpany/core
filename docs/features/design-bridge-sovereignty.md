# Design Doc: Hardware-Attested Bridge Sovereignty (HABS)
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The Azure DevOps MCP bypass (CVE-2026-32211) demonstrated that cloud-to-local bridges often rely on insecure credential inheritance. MCP Any must ensure that tool calls crossing environment boundaries are cryptographically bound to a physical hardware identity to prevent unauthorized access and credential exfiltration.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound (TPM/Secure Enclave) identity tokens for all cross-environment tool calls.
    * Neutralize credential-bypass vulnerabilities in cloud-to-local adapters.
    * Provide a non-repudiable audit trail for bridged tool execution.
* **Non-Goals:**
    * Replacing existing cloud authentication (e.g., OAuth).
    * Managing hardware drivers directly.

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer using Cloud-hosted LLM with Local Git Tools.
* **Primary Goal:** Securely invoke local git commands from a cloud-hosted agent without risking local token exfiltration.
* **The Happy Path (Tasks):**
    1. Agent requests a local tool call via the MCP Any Bridge.
    2. MCP Any requests a hardware-attested identity fragment from the local host.
    3. Host provides a TPM-signed token bound to the current session and mission-root.
    4. MCP Any validates the token and the mission-root lineage.
    5. Tool execution is authorized and logged with hardware provenance.

## 4. Design & Architecture
* **System Flow:**
    [Cloud Agent] -> [MCP Any Bridge] -> [HABS Middleware] -> [Local TPM/SEP] -> [Tool Execution]
* **APIs / Interfaces:**
    * `GET /habs/attest`: Request session attestation.
    * `POST /habs/verify`: Validate cross-environment token.
* **Data Storage/State:**
    * Short-lived, hardware-bound session keys stored in secure kernel memory.

## 5. Alternatives Considered
* **Software-only Tokens (JWTs)**: Rejected because they are susceptible to exfiltration and replay attacks in compromised environments.
* **VPN/Tunneling**: Rejected as a primary solution due to configuration overhead and lack of granular "Per-Call" attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory hardware attestation for every cross-boundary call. No implicit trust for cloud-origin requests.
* **Observability:** Hardware-linked IDs in all audit logs.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
