# Design Doc: Hardware-Bound Identity Hub (HBIH)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The proliferation of AI agents in production environments has exposed a critical weakness in traditional authentication: the reliance on bearer tokens (API keys). These tokens are static strings that can be easily exfiltrated via prompt injection, logic flaws, or accidental exposure. Once leaked, they can be used by any actor from any device, leading to unauthorized resource access and financial loss.

The Hardware-Bound Identity Hub (HBIH) evolves MCP Any from a credential proxy into an authoritative "Hardware Identity Mint." It utilizes device-resident Trusted Platform Modules (TPMs) or Secure Enclaves to issue and verify cryptographic proofs that replace static bearer tokens. This ensures that an agent's identity is physically bound to the authorized hardware, neutralizing the impact of credential exfiltration.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a local identity service that interfaces with TPM 2.0 and Apple Silicon Secure Enclave.
    * Issue time-bound, hardware-attested cryptographic proofs for all outbound agent requests.
    * Provide a standardized interface for external LLM providers and MCP servers to verify these proofs.
    * Enable "Credential Sovereignty" where private keys never leave the device's secure hardware.
* **Non-Goals:**
    * Implementing a full-blown Identity and Access Management (IAM) system (focus is on hardware binding).
    * Supporting legacy hardware without TPM or Secure Enclave capabilities.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an attacker from using a coding agent's Anthropic API key even if they successfully exfiltrate it via prompt injection.
* **The Happy Path (Tasks):**
    1. The architect configures MCP Any to use the HBIH for the "Production Coding Swarm."
    2. The HBIH initializes a device-bound key pair within the local TPM.
    3. An agent attempts to call a code-modifying tool that requires an Anthropic API call.
    4. MCP Any intercepts the request and requests a "Hardware Proof" from the HBIH.
    5. The HBIH signs the request metadata using the TPM-resident private key.
    6. The request is sent to the LLM provider, carrying the hardware-bound signature instead of a raw bearer token.
    7. If an attacker exfiltrates the "API Key" string, it is useless without the device-bound signature from the local HBIH.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[AI Agent] -->|Tool Call| Gateway[MCP Any Gateway]
        Gateway -->|Sign Request| HBIH[HBIH Hub]
        HBIH -->|Sign| TPM[TPM / Secure Enclave]
        TPM -->|Signed Proof| HBIH
        HBIH -->|Proof + Metadata| Gateway
        Gateway -->|Authenticated Call| Upstream[LLM / MCP Server]
        Upstream -->|Verify Proof| Attestation[Attestation Service]
    ```
* **APIs / Interfaces:**
    * `POST /hbih/sign`: Request a cryptographic signature for a specific request payload.
    * `GET /hbih/identity`: Retrieve the public attestation certificate for the device.
    * `POST /hbih/rotate`: Trigger a hardware-locked key rotation.
* **Data Storage/State:**
    * Private keys are stored in the TPM/Secure Enclave (Hardware-bound).
    * Public certificates and metadata are stored in a local encrypted SQLite database.

## 5. Alternatives Considered
* **Short-Lived Tokens (JWT):** Rejected as the primary solution because the initial "Root Token" can still be exfiltrated.
* **IP Whitelisting:** Rejected due to its inability to handle dynamic developer environments and the risk of IP spoofing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The HBIH must itself be protected by local OS-level access controls. Access to the sign endpoint requires a session-bound challenge-response handshake.
* **Observability:** Every signing event is logged in the Hardware Audit Trail with nanosecond precision.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
