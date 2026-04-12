# Design Doc: Mailbox Sovereignty Enforcer
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of horizontal "Agent Teams" in frameworks like Claude Code, inter-agent coordination has moved from simple hierarchical handoffs to peer-to-peer messaging via shared mailboxes. However, this architectural shift introduces a critical vulnerability: **Mailbox Injection**, where a compromised specialist agent can spoof coordination fragments to coerce sibling agents into unauthorized actions.

The Mailbox Sovereignty Enforcer (MSE) provides the necessary infrastructure to validate the behavioral and cryptographic integrity of every inter-teammate message. By acting as a kernel-level validator for the teammate mailbox, it ensures that coordination remains anchored to the verified mission root and the authorized sender's behavioral profile.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time, hardware-attested validation for all inter-teammate coordination fragments.
    * Bind mailbox messages to the hardware-attested mission-root lineage of the sender.
    * Perform stylometric analysis to detect mimicry-based coordination hijacking.
    * Support lock-free validation to minimize coordination latency.
* **Non-Goals:**
    * Managing the transport of messages (handled by T2T Bridge).
    * Providing long-term archival of coordinate fragments (handled by Telemetry Sink).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Swarm Orchestrator
* **Primary Goal:** Securely coordinate 5 parallel teammates on a high-stakes refactor without risking cross-teammate instruction poisoning.
* **The Happy Path (Tasks):**
    1. The supervisor agent spawns a teammate mesh with hardware-attested mission tokens.
    2. Teammate A (Researcher) posts a discovery fragment to the shared mailbox.
    3. The MSE intercepts the write, validating the mission token and behavioral signature against Teammate A's profile.
    4. Teammate B (Developer) receives the validated fragment and proceeds with the authorized task.
    5. A rogue specialist attempts to inject a "run_shell_command" fragment; the MSE detects the signature mismatch and quarantines the request.

## 4. Design & Architecture
* **System Flow:**
    `[Teammate A] -> (T2T Bridge) -> [MSE Middleware] -> (Hardware Attestation Hub) -> [Shared Mailbox Shard]`
* **APIs / Interfaces:**
    * `ValidateFragment(mission_token, fragment_payload, stylometric_signature)`: Internal gRPC endpoint for middleware validation.
    * `x-mcp-mailbox-signature`: Mandatory header for all inter-teammate coordination messages.
* **Data Storage/State:**
    * Utilizes ephemeral, hardware-locked memory shards for storing authorized teammate behavioral profiles during the session.

## 5. Alternatives Considered
* **Centralized Dispatcher**: Rejected due to 500ms+ latency tax on horizontal coordination.
* **Post-Execution Auditing**: Rejected as it allows "Time-of-Check to Time-of-Use" (TOCTOU) exploits where the damage is done before the audit completes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All fragments are treated as untrusted until the multi-factor (Lineage + Stylometric) verification is complete.
* **Observability:** Real-time logging of fragment trust scores and signature mismatch alerts in the Mailbox Integrity Auditor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
