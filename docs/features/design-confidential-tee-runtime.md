# Design Doc: Confidential TEE Runtime Adapter
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The "8,000 Exposed Servers" crisis highlighted the danger of running local command-line tools exposed to AI agents. Even with sandboxing, a compromised host remains a threat. This feature introduces a specialized adapter that executes `CMD` tools inside a Trusted Execution Environment (TEE) like Intel SGX or AWS Nitro Enclaves. This ensures that the code execution is cryptographically isolated from the host OS.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a `tee-cmd` adapter that runs executables in a verified enclave.
    * Implement "Remote Attestation" to prove to the AI agent that the tool is running in a secure environment.
    * Encrypt all I/O between the MCP Any Core and the TEE.
* **Non-Goals:**
    * Running full GUI applications in TEE.
    * Replacing general-purpose Docker sandboxing (TEE is for high-security primitives).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Execute sensitive local scripts (e.g., wallet signing, database migrations) via an AI agent without risk of host-level interception.
* **The Happy Path (Tasks):**
    1. Engineer configures a service with the `tee-cmd` adapter.
    2. MCP Any requests an enclave quote from the hardware.
    3. The AI agent verifies the quote before sending the first command.
    4. Command is executed inside the TEE; results are returned over an encrypted channel.

## 4. Design & Architecture
* **System Flow:**
    `Core Server -> TEE Adapter -> Enclave Manager (e.g., EGo, Occlum) -> Secure Enclave`
* **APIs / Interfaces:**
    * `TeeAdapterInterface`: Extends the standard Upstream interface with `GetAttestationQuote()`.
* **Data Storage/State:**
    * Ephemeral state only. Secrets are injected into the enclave at runtime via secure handoff.

## 5. Alternatives Considered
* **Standard Docker isolation:** Rejected because a root user on the host can still inspect Docker memory. TEE prevents this.
* **Cloud-only HSMs:** Rejected because we need *local* execution for low-latency and air-gapped scenarios.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The enclave itself is a Zero Trust boundary.
* **Observability:** Performance metrics will include "Enclave Entry/Exit" overhead.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
