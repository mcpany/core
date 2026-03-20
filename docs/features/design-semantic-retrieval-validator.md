# Design Doc: Semantic Retrieval Validator (SRV)
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
As AI agents evolve from simple assistants to autonomous operators, they increasingly rely on unstructured datasets to perform complex tasks. Recent market analysis (Vectra AI, Stellar Cyber) reveals a critical vulnerability: **Uncontrolled Retrieval**. Agents often inadvertently extract and output sensitive PII or IP from these datasets in response to seemingly benign queries, or are tricked into summarizing sensitive information for exfiltration via side channels.

The **Semantic Retrieval Validator (SRV)** is designed to act as a governance layer between tool outputs (retrieved context) and the agent's reasoning engine. It ensures that every fragment of retrieved information is semantically aligned with the verified mission intent and free from unauthorized sensitive data before the agent ever "sees" it.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Perform real-time semantic analysis on all tool-retrieved data fragments.
    *   Validate retrieved context against the "Mission Root" intent to detect drift or "Retrieval Smuggling."
    *   Automatically redact or block PII, IP, and hidden imperative instructions (indirect prompt injection) in retrieved data.
    *   Provide hardware-attested logs of all retrieval validation events.
*   **Non-Goals:**
    *   Replacing traditional RBAC/ABAC at the data source (Upstream).
    *   Modifying the underlying LLM's reasoning capability.
    *   Indexing the unstructured data itself (SRV operates on the *result* of retrieval).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Prevent an autonomous "Research Agent" from accidentally exfiltrating customer PII while analyzing public support forums.
*   **The Happy Path (Tasks):**
    1.  The Architect configures a "Mission Root" for the Research Agent: "Summarize public sentiment regarding product X."
    2.  The Agent calls a `search_forum` tool.
    3.  The tool returns a fragment containing a user's accidental post of their credit card number alongside a product complaint.
    4.  The SRV intercepts the output, performs semantic scanning, and identifies the PII.
    5.  The SRV redacts the credit card number and attaches a "PII-Redacted" metadata tag to the fragment.
    6.  The Agent receives the sanitized fragment and continues its summary without exposure to the sensitive data.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Adapter: Call Tool (Search)
        Adapter->>Upstream: Execute Request
        Upstream-->>Adapter: Raw Results
        Adapter->>SRV: Validate Context (Raw Results + Mission Root)
        SRV->>SemanticEngine: Analyze Fragments
        SemanticEngine-->>SRV: Sanitized Fragments + Risk Score
        SRV-->>Adapter: Validated Result
        Adapter-->>Agent: Sanitized Result
    ```
*   **APIs / Interfaces:**
    *   `srv.ValidateContext(fragments []string, missionRoot string) (sanitized []string, auditToken string, err error)`
    *   New configuration block in `config.yaml`:
        ```yaml
        srv:
          enabled: true
          piiStrategy: redact
          intentDriftThreshold: 0.85
          allowedEntities: ["ProductInfo", "PublicSentiment"]
        ```
*   **Data Storage/State:**
    *   SRV utilizes an in-memory cache for mission-root embeddings to speed up semantic comparison.
    *   Audit tokens are persisted in the local SQLite "Blackboard" for forensic review.

## 5. Alternatives Considered
*   **Agent-Side Filtering**: Rejected because a compromised or hallucinating agent cannot be trusted to self-censor its own context.
*   **Static Pattern Matching (RegEx)**: Rejected because it fails to detect "Semantic Retrieval Smuggling" (where the info is not PII but is outside the mission scope).
*   **Source-Layer Filtering**: Often impossible as MCP Any acts as an adapter to third-party APIs where we do not control the retrieval logic.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust)**: SRV is a "Fail-Closed" system. If semantic analysis fails or is inconclusive, the fragment is blocked by default.
*   **Observability**: All redactions and blocked fragments are logged with high-entropy "Retrieval Integrity" alerts in the UI.

## 7. Evolutionary Changelog
*   **2026-03-20:** Initial Document Creation.
