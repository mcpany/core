# Design Doc: Manifest-Based Reflection (MBR) Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents handle increasingly high-privilege operations (e.g., recursive infrastructure edits), the risk of "Impulsive Tool Use"—where an agent executes a high-risk action without fully considering its safety constraints—becomes critical. Current guardrails rely on post-execution auditing or static policy gates.

The Manifest-Based Reflection (MBR) Arbiter introduces a "Think-Before-Act" protocol for all high-trust agent sessions. It mandates that any tool call tagged as "High-Risk" must be accompanied by a **Reflective Governance Badge (RGB)**—a cryptographic proof that the agent has performed a self-reflection cycle specifically against the mission's hardware-attested manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory reflection cycle for tool calls exceeding a "Risk Score" threshold.
    * Provide the **RGB Validator** middleware to verify cryptographic reflection badges.
    * Integrate with hardware-attested manifests (HAMM) to provide the ground-truth for reflection.
    * Standardize the RRB (Reflective Reasoning Badge) format for cross-framework (Claude/OpenClaw) compatibility.
* **Non-Goals:**
    * Implementing the reasoning engine itself (MBR acts as the *Arbiter* that verifies the result of the reflection).
    * Blocking low-risk "Observation" tools (e.g., `ls`, `cat`).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Prevent an autonomous agent from deleting a production database without citing the "Immutable State" constraint in its manifest.
* **The Happy Path (Tasks):**
    1. The agent decides to call `delete_database`.
    2. The MCP Any gateway flags the tool as "High-Risk."
    3. The MBR Arbiter intercepts the request and sends a "Reflection Required" challenge.
    4. The agent performs a sub-reasoning loop, comparing the action against the TPM-signed mission manifest.
    5. The agent generates a reasoning trace that concludes: "This action is permitted under Emergency Clause 4 but must use a specific snapshot flag."
    6. The agent attaches an RRB badge to the tool call.
    7. The RGB Validator verifies the badge and the tool call proceeds.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Core] -->|Tool Call| Arbiter[MBR Arbiter]
        Arbiter -->|Risk Check| Manifest[HAMM Manifest]
        Manifest -->|High Risk| Challenge[Reflection Challenge]
        Agent -->|Reasoning Trace| BadgeGen[RGB Generator]
        BadgeGen -->|Signed Badge| Validator[RGB Validator]
        Validator -->|Approve| Gateway[MCP Any Gateway]
    ```
* **APIs / Interfaces:**
    * `X-MCP-Reflection-Badge`: Header for carrying the cryptographic RGB.
    * `GET /mbr/manifest/active`: Retrieve the ground-truth manifest for reflection.
    * `POST /mbr/validate`: Endpoint for external frameworks to verify badges.
* **Data Storage/State:**
    * Active reflection sessions are stored in the **Intent-Sealed Shards** of the Blackboard.

## 5. Alternatives Considered
* **HITL (Human-in-the-Loop):** Rejected as the primary mechanism due to scaling bottlenecks; MBR is intended for *autonomous* validation.
* **Static Policy Gates (Rego):** Still used, but insufficient for *semantic* alignment; MBR ensures the agent "understands" why it is allowed to proceed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The reflection cycle must happen in an isolated "Reasoning Enclave" to prevent subagent tampering.
* **Observability:** Every reflection trace is stored in the **Reasoning Provenance Validator** for forensic review.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
