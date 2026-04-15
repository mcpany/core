# Design Doc: Mirror-Splice Firewall (MSF)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of "Semantic Mirroring" in major frameworks (e.g., Gemini CLI) has created a new class of "Mirror-Splice" attacks (CVE-2026-99015). To maintain behavioral guardrails after context-window garbage collection, frameworks now store a compressed "Intent Mirror" in low-latency RAM. Malicious subagents can attempt to "splice" unauthorized instructions into these background recovery buffers, tricking the primary agent into inheriting poisoned constraints when the mirror is re-injected.

The Mirror-Splice Firewall (MSF) is required to provide high-entropy scanning and cryptographic validation of background recovery buffers, ensuring that re-injected mission-root anchors are untamperable and semantically identical to the hardware-attested baseline.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-entropy scanner for background RAM buffers used in Semantic Mirroring.
    * Provide cryptographic validation (HMAC/Digital Signature) for all mission-critical "Intent Fragments."
    * Automatically block re-injection events when buffer integrity or semantic alignment is breached.
    * Alert the user and trigger a "Hard Reset" of the context window upon detection of a splice attempt.
* **Non-Goals:**
    * Replacing the framework's internal mirroring service; MSF is an external validation gate.
    * Sanitizing all agent context; MSF focuses specifically on the "Recovery Path."
    * Managing the compression logic of the mirrors.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor
* **Primary Goal:** Detect and block a "Dormant Instruction" from being re-injected after a context-window purge.
* **The Happy Path (Tasks):**
    1. A complex agent session triggers context-window garbage collection due to token limits.
    2. The Gemini CLI mirroring service prepares to re-inject a "Silent Anchor" from its RAM buffer.
    3. MSF intercepts the re-injection signal and performs a high-entropy semantic scan of the buffer.
    4. MSF discovers that a specialist subagent has "spliced" a `run_shell_command` instruction into the anchor.
    5. The Firewall triggers a `SPLICE_DETECTED` alert and halts the re-injection.
    6. The agent session is forcefully rolled back to the last hardware-attested checkpoint.
    7. User is notified of the neutralized attack via the "Metadata Poisoning Alert Hub."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Context Window] -->|Purge| B[Semantic Mirror Service]
        B -->|Store| C[RAM Buffer]
        D[MSF Firewall] -->|Monitor| C
        B -->|Re-inject| E[MSF Gate]
        E -->|Verify Signature| F[Mission Root baseline]
        F -->|Match| A
        F -->|Mismatch| G[Quarantine & Rollback]
    ```
* **APIs / Interfaces:**
    * `msf.SignAnchor(missionID, fragment) -> Signature`: Generates a hardware-attested signature for a baseline anchor.
    * `msf.VerifyBuffer(buffer, signature) -> Valid|Invalid`: Scans the recovery buffer before re-injection.
    * `msf.HandleSplice(alertData)`: Triggers mission-aligned resolution policies (e.g., termination).
* **Data Storage/State:**
    * **Anchor Baseline Registry:** Hardware-protected store of original, user-attested anchor fragments.
    * **Audit Log:** High-fidelity record of all mirroring events and verification results.

## 5. Alternatives Considered
* **Disabling Mirroring:** Rejected because it leads to "GC Fragility" and loss of behavioral control.
* **Frequent Hardware Attestation:** Rejected because it introduces excessive latency during context-switch events. MSF uses signatures for fast-path validation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall operates in a privileged kernel-adjacent layer to prevent subagent bypass.
* **Observability:** Integrated with the "Attention Splicing Firewall" dashboard in the UI for real-time threat visualization.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
