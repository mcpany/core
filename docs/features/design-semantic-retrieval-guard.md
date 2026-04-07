# Design Doc: Semantic Retrieval Guard (SRG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents increasingly handle large, unstructured datasets (e.g., via RAG), the risk of "Uncontrolled Retrieval" has emerged as a primary security threat. Agents often inadvertently retrieve and output Personally Identifiable Information (PII) or Intellectual Property (IP) that they are authorized to *read* but the end-user (or a specialist subagent) is not authorized to *see*.

The Semantic Retrieval Guard (SRG) is a mandatory middleware layer in MCP Any that performs real-time, intent-aware scanning of all retrieved context fragments before they are injected into the agent reasoning loop.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all tool outputs and RAG retrieval fragments.
    * Perform real-time PII/IP identification using semantic analysis.
    * Validate retrieval fragments against the cryptographically signed "Mission Root" intent.
    * Redact or block sensitive fragments that exceed the authorized scope.
* **Non-Goals:**
    * Replacing general-purpose Data Loss Prevention (DLP) systems; it is agent-specific.
    * Modifying the underlying data sources; it only sanitizes the retrieved *view*.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Administrator
* **Primary Goal:** Prevent an HR agent from leaking employee salaries when answering a general "What is the department structure?" query.
* **The Happy Path (Tasks):**
    1. Agent invokes a `retrieve_docs` tool for department info.
    2. The tool returns a JSON fragment containing names, titles, and accidentally includes salary fields.
    3. SRG intercepts the fragment and identifies "salary" as sensitive IP/PII.
    4. SRG checks the Mission-Root (which only authorizes "structural hierarchy").
    5. SRG redacts the salary field from the fragment.
    6. The sanitized fragment is passed to the agent's attention window.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[MCP Tool] -->|Raw Fragment| B[SRG Middleware]
        B --> C{Intent Match?}
        C -->|No| D[Redact/Block]
        C -->|Yes| E[Scan for PII]
        E --> F{PII Found?}
        F -->|Yes| G[Mask/Redact]
        F -->|No| H[Inject into Context]
        G --> H
    ```
* **APIs / Interfaces:**
    * `srg.Sanitize(fragment, missionToken) -> SanitizedFragment`: The core entry point for sanitization.
* **Data Storage/State:**
    * **Policy Store:** Local cache of redaction rules and sensitive patterns.

## 5. Alternatives Considered
* **Static Keyword Filtering:** Rejected because it fails against semantic context (e.g., masking "compensation" but missing "yearly total").
* **Source-Side Access Control:** Ideal, but often impossible when agents access unstructured legacy data. SRG provides a "Last Mile" defense.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SRG itself runs in an isolated WASM sandbox to prevent "Scrubbing-as-an-Exploit" where the sanitizer is hijacked.
* **Observability:** Integrated with the "IDS Status Monitor" for real-time visualization of redacted fragments.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
