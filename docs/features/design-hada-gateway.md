# Design Doc: Hardware-Attested Discovery Approval (HADA) Gateway
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Gemini CLI and other agents have been vulnerable to "Pre-Flight" RCE where untrusted workspace settings for tool discovery (e.g., `tools.discoveryCommand`) execute automatically. HADA ensures that no discovery-time execution occurs without an explicit, hardware-attested approval from the user.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate TPM/Secure Enclave-bound approval for all discovery commands.
    * Prevent "Unknown Trust" from defaulting to execution.
    * Provide a cryptographically signed "Safe-to-Discover" token for the sandbox.
* **Non-Goals:**
    * Managing tool-call approvals (handled by HITL Middleware).
    * Providing general-purpose hardware attestation.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Agent User
* **Primary Goal:** Prevent arbitrary code execution when running an agent in a new directory.
* **The Happy Path (Tasks):**
    1. User runs `mcpany` in a directory with a malicious `.gemini/settings.json`.
    2. HADA intercepts the `discoveryCommand` request.
    3. HADA triggers a hardware-attestation prompt on the user's device (e.g., TouchID/Windows Hello).
    4. User reviews the specific command in the HADA UI.
    5. Upon approval, HADA generates a TPM-signed token and allows execution in the Pre-Flight Trust Enclave.

## 4. Design & Architecture
* **System Flow:**
    `Discovery Trigger -> HADA Interceptor -> Hardware Challenge -> TPM Signing -> Enclave Execution`
* **APIs / Interfaces:**
    * `hada.ChallengeUser(commandSummary) -> AttestationToken`
    * `hada.VerifyApproval(token) -> boolean`
* **Data Storage/State:**
    * **Approved Command Cache:** Local database of hashes for previously approved discovery commands.

## 5. Alternatives Considered
* **Trust Prompts (Software-only):** Rejected because they are susceptible to clickjacking and programmatic bypass. Hardware-binding is required for absolute sovereignty.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leverages local TPM for non-repudiable approval.
* **Observability:** Logs all approval events to the Local Security Audit Service.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
