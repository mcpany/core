# Design Doc: Relational Proof-of-Intent (PoI) Validator
**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
Current security models for AI agent swarms rely heavily on identity attestation (e.g., "Is this Subagent A?"). However, market sync research from 2026-03-24 reveals the rise of "Context-Mirroring" attacks (CVE-2026-34015), where a legitimate subagent is coerced into executing actions that diverge from the parent's mission but remain within the agent's technical permissions.

MCP Any needs to solve this by moving from **Identity-Bound Trust** to **Relational Intent-Bound Trust**. This design introduces a validator that verifies not just *who* is calling a tool, but *why*—by checking the tool call against a cryptographically signed chain of intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement UACO v1.7 PoI header validation.
    * Verify the complete "Intent Chain" from the Mission Root to the current Subagent.
    * Neutralize Context-Mirroring by blocking tool calls that lack a verifiable lineage.
    * Support hardware-attested reasoning steps (Gemini-style provenance).
* **Non-Goals:**
    * Implementing the models' reasoning logic itself.
    * Replacing the existing Policy Firewall (this acts as a pre-filter).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Code Reviewer" subagent from being tricked into "exfiltrating" the codebase to a malicious URL using its read permissions.
* **The Happy Path (Tasks):**
    1. The Mission Root signs an intent: "Review PR #123".
    2. The Parent Agent spawns a Subagent with a derived intent: "Read files in PR #123".
    3. The Subagent attempts to call `fs.read_file`.
    4. **Relational PoI Validator** intercepts the call.
    5. Validator checks the PoI header: it contains a signature from the Parent Agent, which is linked to the Mission Root.
    6. Validator confirms `fs.read_file` aligns with the reviewer's derived intent.
    7. Tool call is permitted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Tool Call + PoI Header| Gateway[MCP Any Gateway]
        Gateway -->|Verify Chain| PoIValidator[Relational PoI Validator]
        PoIValidator -->|Query Root| MissionRegistry[Mission Root Registry]
        PoIValidator -->|Check Signatures| HSM[Hardware Security Module / TPM]
        PoIValidator -->|Result| Firewall[Policy Firewall]
    ```
* **APIs / Interfaces:**
    * `X-UACO-Intent-Chain`: A base64 encoded JWT-like structure containing signed intent fragments.
    * `VerifyIntentChain(chain []byte) (bool, error)`: Internal service method.
* **Data Storage/State:**
    * Mission Root hashes are stored in the `MissionRegistry` (ephemeral/session-bound).

## 5. Alternatives Considered
* **Flat Intent Tokens:** Rejected because they don't prevent subagents from "re-purposing" a parent token for a sibling's task.
* **Pure Behavioral Monitoring:** Rejected as a primary defense due to high false-positive rates; MSIV (Multi-Step Intent Verification) is more deterministic.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All intent fragments must be signed by a hardware-attested key.
* **Observability:** Audit logs will include the "Intent Lineage" for every blocked or allowed call.

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
