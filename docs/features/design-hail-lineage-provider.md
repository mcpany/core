<!-- markdownlint-disable MD013 MD030 MD032 MD022 MD007 MD033 MD031 MD004 MD024 MD026 MD012 MD003 MD029 MD040 MD009 -->
# Design Doc: Hardware-Attested Intent Lineage (HAIL)
**Status:** Draft
**Created:** [2026-06-19]

## 1. Context and Scope
The emergence of **Reasoning Path Shadowing** (stylometric mimicry) has exposed a critical gap in multi-agent swarms. In this attack, a malicious specialist agent mimics the "Stylometric Signature" and "Chain-of-Thought" structure of a parent agent to inject instructions that pass standard consistency checks. MCP Any needs to provide **Hardware-Attested Intent Lineage (HAIL)** to cryptographically link every reasoning fragment back to a TPM-signed mission root, ensuring that the "author" of any instruction is non-repudiable and verified.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Generate cryptographically signed "Reasoning Fragments" for all inter-agent messages.
    *   Link every fragment back to a hardware-attested "Mission Root."
    *   Provide real-time stylometric analysis to verify the author of reasoning fragments (defense against mimicry).
    *   Maintain an immutable, hardware-locked audit trail of the reasoning path.
*   **Non-Goals:**
    *   Eliminating all model-level hallucinations (we verify the *lineage*, not the *truth*).
    *   Protecting against local-user-initiated intent shifts (the user is the root of trust).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Verify that all code changes proposed by a multi-agent swarm genuinely originate from the user-authorized mission root, and haven't been "shadowed" by a malicious subagent.
*   **The Happy Path (Tasks):**
    1.  The User initiates a mission via the MCP Any Gateway.
    2.  MCP Any mints a hardware-attested **HAIL Mission Root Token**.
    3.  Every tool call or sub-mission spawned by the lead agent must include a **HAIL Child Fragment**, signed by the lead agent's identity and linked to the root token.
    4.  A specialist agent attempts to "Shadow" the parent agent's style to escalate its permissions.
    5.  The **Stylometric Verification Hub** detects a mismatch between the reasoning style and the hardware-attested identity.
    6.  MCP Any blocks the tool call and alerts the user of a lineage violation.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Token] --> B[SRM Provider]
        B --> C{HAIL Token Minting}
        C --> D[Hardware-Signed Fragment]
        E[Subagent Reasoning Trace] --> F[Stylometric Hub]
        F --> G[Signature Verification]
        G -- Match --> H[Authenticated Instruction]
        G -- Mismatch --> I[Lineage Alert / Block]
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/lineage/sign`: Accept a reasoning fragment and return a HAIL-signed token.
    *   `GET /v1/lineage/verify`: Verify the lineage and stylometric signature of a fragment.
*   **Data Storage/State:**
    *   Lineage tokens are stored in the **Mesh-Resident Lineage Tracker** (forensic audit log).

## 5. Alternatives Considered
*   **Token-only Lineage:** Rejected because tokens can be "Shadowed" if the model's output is not cryptographically bound to the hardware session.
*   **Manual Code Review:** Rejected because it cannot keep pace with machine-speed swarm coordination and doesn't address the stylometric mimicry threat.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The HAIL Provider is the authoritative root of cognitive trust; its private keys are never exposed and remain TPM-bound.
*   **Observability:** The "Lineage Inspector" UI will provide a visual "Chain of Command" for every tool call in the swarm.

## 7. Evolutionary Changelog
*   **[2026-06-19]:** Initial Document Creation.
