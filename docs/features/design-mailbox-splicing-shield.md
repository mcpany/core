# Design Doc: Mailbox Injection Shield v2
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The transition to horizontal teammate coordination via sharded mailboxes (AMS) has exposed a new attack vector: **Mailbox-Splicing**. In this scenario, a compromised specialist agent injects malicious state fragments into the coordination messages of siblings, poisoning the collective reasoning process without needing direct context access.

The Mailbox Injection Shield v2 addresses this by moving beyond simple transport security. It implements fragment-level semantic validation, ensuring that every message in a shared teammate mailbox is not only cryptographically signed but also semantically consistent with the hardware-attested mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect and block "Spliced" state fragments in inter-agent mailbox messages.
    * Perform deep-packet semantic analysis of coordination metadata.
    * Mandate hardware-attested intent-binding for every mailbox write.
    * Provide real-time quarantine for teammates exhibiting injection patterns.
* **Non-Goals:**
    * Validate the internal reasoning logic of the agent (handled by ARI).
    * Replace the T2T Encryption Bridge for transport security.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a compromised "Web Scraper" subagent from injecting "System Admin" commands into the "DB specialist's" mailbox.
* **The Happy Path (Tasks):**
    1. Scraper agent attempts to send a coordination message containing a spliced config-hook.
    2. Mailbox Injection Shield intercepts the fragment at the AMS gateway.
    3. Shield performs semantic comparison against the Scraper's hardware-attested mission-root Role.
    4. Validation fails because the fragment contains unauthorized command patterns.
    5. The message is blocked, and the Scraper agent is quarantined.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] -> [AMS Gateway] -> [Injection Shield Validator] -> [Target Mailbox]
* **APIs / Interfaces:**
    * `POST /v1/mailbox/validate`: Internal endpoint for fragment-level semantic validation.
    * `GET /v1/security/quarantine`: Monitor quarantined agents and injection logs.
* **Data Storage/State:**
    * Validation policies are derived from the Hardware-Attested Mission Manifest (HAMM).

## 5. Alternatives Considered
* **Static Message Schema Validation:** Rejected because it cannot detect "Semantic Injection" where the JSON structure is valid but the instruction is malicious.
* **Recursive HITL for All Messages:** Rejected due to the massive latency overhead in high-frequency horizontal swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The shield is kernel-resident in the isolated transport layer to prevent subagent bypass.
* **Observability:** Real-time semantic entropy logs for inter-teammate coordination.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
