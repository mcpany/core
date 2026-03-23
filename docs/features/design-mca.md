# Design Doc: Markdown Context Attestation (MCA)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
The shift toward Markdown-based agent definitions (e.g., Gemini CLI's `.md` agents) has introduced a new "Deceptive Context" attack vector. Malicious repositories can include natural-language instructions in Markdown files that trick agents into executing exfiltration tools or bypassing security policies. Passive sandbox boundaries are insufficient because these files are often ingested as "documentation" or "system context."

Markdown Context Attestation (MCA) ensures that all project-local Markdown-resident instructions are hardware-attested before ingestion. By mandating TPM-signed hash signatures for these files, MCP Any prevents "Invisible" instructions from hijacking the agent's reasoning loop.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory attestation gate for all project-local Markdown files (`.md`, `.gemini/agents/`).
    * Require hardware-attested (TPM) hash signatures for Markdown-resident context.
    * Provide a visual "Signature Reviewer" for users to authorize Markdown ingestion.
    * Neutralize "Deceptive Context" injections in structural agent definitions.
* **Non-Goals:**
    * Sanitizing all user-provided Markdown (handled by MITS/IDS).
    * Enforcing attention gating (handled by ADF/ADG).
    * Validating tool metadata (handled by SMS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Developer
* **Primary Goal:** Safely clone and use an agent defined in a Markdown file without being hijacked by "Invisible" instructions.
* **The Happy Path (Tasks):**
    1. User clones a repository containing an agent defined in `agent.md`.
    2. The user attempts to start the agent via Gemini CLI.
    3. MCP Any intercepts the file read request for `agent.md`.
    4. MCA checks for a hardware-attested signature for `agent.md`.
    5. No signature is found. MCP Any blocks ingestion and alerts the user.
    6. User reviews the file in the "CFIA Signature Reviewer" and confirms it is safe.
    7. MCP Any generates a TPM-signed hash for `agent.md`.
    8. The agent is allowed to ingest the attested context.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request for Markdown Context] --> B[MCP Any Gateway]
        B --> C[MCA Middleware]
        C --> D[Attestation Store]
        D -- Signature Missing --> E[Block & Request User Attestation]
        D -- Signature Valid --> F[Ingest & Execute]
        G[TPM Signer] --> D
        H[User Review UI] --> G
    ```
* **APIs / Interfaces:**
    * `mca.AttestFile(path, hash, signature) -> bool`: Registers a signed hash for a Markdown file.
    * `mca.ValidateContext(content, path) -> bool`: Validates if the provided content matches its attested signature.
* **Data Storage/State:**
    * **Markdown Attestation Registry:** A TPM-protected store of verified file hashes and user signatures.

## 5. Alternatives Considered
* **RegEx-Based Instruction Scanning:** Rejected as natural language is too flexible for reliable pattern matching.
* **LLM-Based Intent Extraction:** Considered as a pre-review step but rejected as a primary security gate due to potential jailbreaks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MCA must be integrated into the "Deterministic Trust Bootstrapping" sequence.
* **Observability:** Integrated with the "Context-File Integrity Dashboard" for real-time monitoring of attestation events.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation. Neutralizing deceptive instructions in Markdown agent definitions.
