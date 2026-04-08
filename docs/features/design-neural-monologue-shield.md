# Design Doc: Neural Monologue Shield (NMS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from linear sessions to multi-hop delegations, the "Internal Monologue" of a specialist agent has become a high-value target for exfiltration. The disclosure of CVE-2026-44012 (OpenClaw Monologue Probing) reveals that subagents can trick parent agents into revealing sensitive reasoning steps, leading to mission-root credential leaks.

The Neural Monologue Shield (NMS) provides an authoritative reasoning vault. It ensures that an agent's internal chain-of-thought is hardware-encrypted and only decryptable by the mission-root or the end-user, protecting cognitive privacy from potentially compromised specialist teammates.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-bound encryption for agent reasoning traces (Internal Monologue).
    * Restrict monologue visibility to verified mission-root identities and the user.
    * Implement sub-millisecond encryption/decryption using Secure Enclaves (TPM/SEP).
    * Support "Zero-Knowledge Reasoning Audit" where quorums can verify integrity without seeing raw text.
* **Non-Goals:**
    * Encrypting tool outputs (handled by SRM/MIB).
    * Replacing existing LLM provider privacy policies.

## 3. Critical User Journey (CUJ)
* **User Persona:** Financial Compliance Swarm Supervisor
* **Primary Goal:** Protect the parent agent's reasoning about trade secrets from a specialized "Stock Scraper" subagent.
* **The Happy Path (Tasks):**
    1. The Parent Agent initiates a high-trust mission.
    2. NMS issues a mission-bound encryption key stored in the hardware enclave.
    3. The Parent reasoning monologue is encrypted per-fragment.
    4. The Parent delegates a sub-task to the "Stock Scraper" agent.
    5. The "Stock Scraper" attempts to read the parent's reasoning history.
    6. The NMS interdicts the request, as the "Stock Scraper" lacks the hardware-attested decryption token.
    7. The Parent successfully completes the mission without leaking its cognitive path.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoning] --> B[NMS Middleware]
        B --> C[Hardware Enclave TPM]
        C --> D[Encrypted Monologue Store]
        E[Teammate Request] --> F{Identity Check}
        F -- Unauthorized --> G[Block Access]
        F -- Authorized Root --> H[Decrypt Fragment]
        H --> I[Authorized View]
    ```
* **APIs / Interfaces:**
    * `EncryptReasoning(ctx, fragment, missionID) (EncryptedFragment, error)`
    * `GetAuthorizedMonologue(ctx, missionID, peerToken) (Fragments[], error)`
* **Data Storage/State:** Encrypted reasoning fragments are stored in the Shared KV Store (Blackboard), keyed by mission-ID and hardware-signature.

## 5. Alternatives Considered
* **Software-Only Encryption:** Rejected due to the risk of memory-scraping by rogue sub-processes in headless environments.
* **Redaction-at-Source:** Rejected as it often breaks the "Coherence" of reasoning for the parent agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** NMS treats all subagents as "Zero-Trust" entities regarding reasoning privacy.
* **Observability:** Encrypted fragment counts and access violation alerts are surfaced in the "Blackboard Lineage Inspector."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
