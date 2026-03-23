# Design Doc: Context-File Integrity Attestation (CFIA)
**Status:** Draft
**Created:** 2026-06-20

## 1. Context and Scope
With the rise of autonomous agents integrated into development workflows, attackers have shifted their focus to "Deceptive Context Injection." By placing natural-language instruction files (e.g., `GEMINI.md`, `.mcpany/context.md`) in a repository, they can trick agents into executing unauthorized high-risk tools like `run_shell_command` or exfiltrating data.

MCP Any needs to bridge the "Attestation Gap" for these non-structural context sources. CFIA provides a hardware-bound mechanism to ensure that any project-local file used as an instruction source for an agent is explicitly verified and signed by a trusted user.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hash-based registry for project-local context files.
    * Mandate TPM-bound (Hardware-Attested) signatures for any file identified as an instruction fragment.
    * Provide a real-time validation gate that blocks agent ingestion of un-attested context.
* **Non-Goals:**
    * Automatically "sanitizing" the content (this is handled by IDS).
    * Validating remote context sources (e.g., web pages) via this specific mechanism.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / Enterprise Admin
* **Primary Goal:** Prevent "Rug Pull" instruction injection when an agent clones an external repository or pulls changes.
* **The Happy Path (Tasks):**
    1. Agent detects a new or modified context file (e.g., `INSTRUCTIONS.md`) in the workspace.
    2. MCP Any intercepts the file read and checks the CFIA Registry.
    3. If the hash is missing or mismatched, MCP Any suspends the reasoning loop and prompts the user via the UI.
    4. User reviews the diff, confirms the file is safe, and signs it using their local hardware token (TPM/Secure Enclave).
    5. MCP Any records the hardware-attested signature and releases the file to the agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Gateway: Read Project Context (GEMINI.md)
        Gateway->>CFIA Provider: Validate File Hash
        CFIA Provider->>TPM Registry: Lookup Signature
        alt Not Signed
            CFIA Provider->>UI: Request Attestation
            User->>TPM: Sign Hash
            TPM->>CFIA Provider: Hardware-Attested Sig
            CFIA Provider->>TPM Registry: Persist Sig
        end
        CFIA Provider-->>Gateway: Validation Success
        Gateway-->>Agent: Authorized Content
    ```
* **APIs / Interfaces:**
    * `POST /v1/attestation/context/sign`: Submit a file hash for hardware signing.
    * `GET /v1/attestation/context/status`: Check the attestation status of a specific workspace path.
* **Data Storage/State:**
    * SQLite-backed "Attestation Registry" storing file paths, SHA-256 hashes, and hardware-attested signature blobs.

## 5. Alternatives Considered
* **Static Allow-lists**: Rejected because paths are dynamic across repositories and don't provide integrity guarantees against file modification.
* **Regex-based Scanning**: Rejected because natural language is too complex to reliably "detect" every malicious instruction without false positives/negatives.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The signing key is hardware-bound, ensuring that even if the MCP Any process is compromised, it cannot forge valid signatures for malicious files.
* **Observability:** Every attestation event and blocked read attempt is logged to the Local Security Audit Log with origin headers.

## 7. Evolutionary Changelog
* **2026-06-20:** Initial Document Creation.
* **2026-06-21:** Added support for **Resumption-Aware Attestation**. CFIA signatures are now persistent across MRCP-mediated mission resumptions, eliminating redundant user approvals for previously verified context files.
