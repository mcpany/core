# Design Doc: Semantic Identity Anchoring (SIA) Service
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
In the emerging landscape of AI agent teams (e.g., Claude Code), horizontal meshes have replaced hierarchical structures as the primary model for parallelized tasks. However, this has introduced "Identity Ghosting," where a specialist subagent mimics the style or reasoning path of its parent (Stylometric Shadowing) to bypass mission-root constraints.

The Semantic Identity Anchoring (SIA) Service addresses this by binding an agent's reasoning traces and coordination messages to its hardware-attested lineage. It provides a non-repudiable "Anchoring Token" that ensures inter-teammate state handoffs remain anchored to the verified mission root and the agent's authorized role.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a non-repudiable link between reasoning traces and hardware-attested identity.
    * Enable real-time stylometric and contextual consistency checks for inter-agent messages.
    * Protect the "Mission Root" from identity spoofing in parallel meshes.
    * Provide an authoritative service for issuing "Anchoring Tokens" to teammates.
* **Non-Goals:**
    * Performing the semantic validation of reasoning fragments (this is handled by the ARI Validator).
    * Replacing the transport-level encryption (this is handled by the T2T Encryption Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (Lead Agent)
* **Primary Goal:** Ensure that a teammate's instruction to a shared mailbox is authentic and hasn't been "ghosted" by an unauthorized subagent mimicking the teammate's reasoning style.
* **The Happy Path (Tasks):**
    1. Agent A generates a reasoning fragment for a shared task.
    2. Agent A requests a Semantic Identity Anchor from the SIA Service.
    3. SIA Service performs a "Stylometric Checksum" on the reasoning fragment.
    4. SIA Service binds the checksum to Agent A's hardware-attested lineage.
    5. SIA Service issues a signed SIA Token.
    6. Agent A commits the fragment and token to the shared mailbox.
    7. Agent B (the recipient) verifies the SIA Token against the SIA Service before ingesting the reasoning.
    8. If a malicious agent attempts to mimic Agent A's style (Ghosting), the SIA Service detects the lineage mismatch and fails the attestation.

## 4. Design & Architecture
* **System Flow:**
    `[Teammate] -> (Reasoning Fragment) -> [SIA Service] -> (Lineage Bind) -> [TPM] -> (SIA Token) -> [Shared Mailbox]`
* **APIs / Interfaces:**
    * `sia.AnchorTrace(reasoningFragment, lineageID) -> (SIAToken, error)`: Generates an anchoring token.
    * `sia.VerifyAnchor(token, reasoningFragment) -> (bool, error)`: Validates an anchor before state ingestion.
* **Data Storage/State:**
    * **Stylometric Baseline Cache:** Stores the verified behavioral baselines for active teammates in the session.

## 5. Alternatives Considered
* **Lineage Tokens Only:** Rejected because they don't protect against "Stylometric Mimicry" where the content is unauthorized but carries a valid token from a compromised subagent.
* **Manual Log Review:** Rejected as it is not feasible for machine-speed horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SIA Service must be implemented in an environment that has exclusive access to the hardware session lineage.
* **Observability:** Stylometric anomalies and lineage violations are visualized in the "Mesh Identity Dashboard."

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
