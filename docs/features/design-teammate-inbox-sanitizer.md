# Design Doc: Teammate Inbox Sanitizer (TIS)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
The transition to parallel Agent Teams (Claude Opus 4.6) introduces a new attack vector: "Inbox Injection." Malicious instructions can be spliced into inter-teammate coordination channels, leading to "Intent Drift" or unauthorized actions by autonomous specialists. TIS provides real-time semantic deconstruction of inbox messages to ensure that all delegated tasks remain strictly aligned with the parent mission-root manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic validation of all messages in the inter-agent inbox.
    * Detect and block "Instruction Splicing" where subagents attempt to override parent constraints.
    * Maintain a verifiable hash-chain of all coordinate fragments.
* **Non-Goals:**
    * Replacing existing transport-layer security (mTLS).
    * Providing long-term message storage (handled by A2A Mailbox).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Manager
* **Primary Goal:** Ensure a specialized "Researcher" agent cannot trick a "Writer" agent into exfiltrating data via a spliced coordination message.
* **The Happy Path (Tasks):**
    1. Researcher agent posts a coordination message to the shared Inbox.
    2. TIS intercepts the message before it is visible to the Writer agent.
    3. TIS deconstructs the message and compares the embedded instructions against the mission-root manifest.
    4. TIS verifies the hash-chain integrity of the coordinating fragments.
    5. If aligned, the message is released to the Writer's inbox; if malicious, the message is quarantined and an alert is triggered.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A (Sender)] -> [Inbox Hub] -> [TIS Interceptor (Semantic Scan)] -> [Agent B (Recipient)]`
* **APIs / Interfaces:**
    * `X-Teammate-Intent-Hash`: Header for coordinating fragment lineage.
    * `inbox.Sanitize(message)`: Internal middleware interface for message deconstruction.
* **Data Storage/State:**
    * Temporary session-bound fragment registry for hash-chain verification.

## 5. Alternatives Considered
* **Binary Allow-listing:** Rejected as it fails to capture natural-language instruction drift.
* **Manual Review:** Rejected due to the "Coordination Stall" in high-speed parallel swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enforces fragment-level sovereignty. Prevents subagent privilege escalation via coordination side-channels.
* **Observability:** Logs all sanitized and quarantined messages in the Mission-Root Audit Trail.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
