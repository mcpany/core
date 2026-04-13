# Design Doc: Semantic Egress Filtering (SEF)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents become more autonomous, they increasingly perform retrieval and summarization tasks across large, unstructured datasets. A critical risk has emerged where agents inadvertently leak Personally Identifiable Information (PII) or Intellectual Property (IP) by including it in summaries intended for lower-clearance users or external communication channels (e.g., Slack, Discord). This "Uncontrolled Retrieval" and subsequent exfiltration through side-channel summarization bypasses traditional access controls.

MCP Any needs a "Semantic Firewall" to intercept agent outputs and perform real-time, high-entropy analysis to detect and redact sensitive data before it crosses the egress boundary.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic analysis of agent-generated text (summaries, reports, messages).
    * Detect and redact PII (names, emails, SSNs) and IP (proprietary code, trade secrets) based on mission-root security policies.
    * Enforce clearance-level awareness for egress channels.
* **Non-Goals:**
    * Replacing existing Data Loss Prevention (DLP) systems (will act as an agent-native complement).
    * Encrypting data at rest (handled by other components).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Data Compliance Officer
* **Primary Goal:** Prevent accidental leakage of proprietary source code via agent-generated summaries shared on public Slack channels.
* **The Happy Path (Tasks):**
    1. Compliance Officer defines a "Knowledge Manifest" and "Redaction Policy" for the development mission.
    2. Agent retrieves multiple files from the repository and generates a summary of the project architecture.
    3. The agent attempts to post this summary to a team Slack channel.
    4. SEF intercepts the outgoing message and performs a semantic scan.
    5. SEF identifies a fragment of proprietary authentication logic in the summary.
    6. SEF redacts the sensitive fragment, replacing it with `[REDACTED - PROPRIETARY CODE]`.
    7. The sanitized summary is delivered to Slack, and an alert is logged in the Mission Audit Trail.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Output] -> [Egress Bridge] -> [SEF Engine (Semantic Scanner)] -> [Redaction Middleware] -> [Target Channel]`
* **APIs / Interfaces:**
    * `POST /egress/filter`: Internal endpoint used by bridges to submit text for filtering.
    * `X-MCP-Clearance-Level`: Header indicating the trust level of the target egress channel.
* **Data Storage/State:**
    * Policy definitions and knowledge manifests are stored in the TPM-bound mission configuration.

## 5. Alternatives Considered
* **Keyword-based Filtering:** Rejected because it fails to capture context-dependent IP leaks and is easily bypassed by paraphrasing.
* **Full Human Review:** Rejected due to the "Manual Review" bottleneck and the speed of autonomous swarm operations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SEF acts as a mandatory gate for all non-local egress. It uses hardware-attested policies to prevent agent-level overrides.
* **Observability:** Logs all redaction events with "Confidence Scores" and maps them to the reasoning trace for forensic analysis.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
