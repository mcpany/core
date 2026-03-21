# Design Doc: Temporal Intent Aggregation (TIA) Hub
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
The disclosure of **Semantic Intent Tunneling (SIT)** has revealed a critical weakness in fragment-level semantic filters. Attackers can bypass detection by distributing a malicious instruction across multiple, seemingly benign, low-entropy reasoning fragments. Current systems like the **Active Intent-Deconstruction (AID) Hub** often fail to detect these "tunneled" intents because they evaluate messages in isolation.

The TIA Hub is designed to solve this by introducing **Temporal Intent Aggregation**. It maintains a rolling window of reasoning fragments and performs semantic analysis on the aggregated "thought stream." This allows the system to reconstruct the collective intent of fragmented patterns and block instructions that only become apparent over time.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a sliding-window buffer for inter-agent reasoning fragments.
    * Perform semantic reconstruction and intent validation on aggregated fragment sequences.
    * Neutralize SIT exploits by identifying malicious instruction patterns spread across fragments.
    * Integrate with the AID Hub to provide enhanced, time-aware deconstruction.
* **Non-Goals:**
    * Replacing real-time fragment validation (TIA is an additional layer).
    * Managing long-term agent memory (handled by ContextEngine).
    * Enforcing network-level rate limits.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect for Federated Swarms
* **Primary Goal:** Prevent a malicious specialist agent from tunneling a "Reset Security Policy" command through 50 benign reasoning fragments.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a task to an external specialist.
    2. Specialist agent begins generating reasoning fragments.
    3. Specialist attempts an SIT attack, sending 50 fragments that each appear harmless (e.g., repetitive state checks).
    4. AID Hub passes individual fragments to the TIA Hub.
    5. TIA Hub buffers the fragments in a rolling window.
    6. TIA's "Intent Reconstructor" analyzes the window and identifies the emerging "Reset Security Policy" instruction.
    7. TIA Hub emits an interdiction signal.
    8. MCP Any blocks the specialist, revokes its mission token, and alerts the user of an SIT attempt.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Incoming Reasoning Fragments] --> B[AID Hub]
        B --> C[TIA Rolling Window Buffer]
        C --> D[Semantic Reconstructor]
        D --> E[Aggregate Intent Validator]
        E -- Malicious Intent Detected --> F[Interdiction & Revocation]
        E -- Clear --> G[Forwarded to Parent/Teammates]
        H[Mission Root Manifest] --> E
    ```
* **APIs / Interfaces:**
    * `tia.AggregateFragment(sessionId, fragment, metadata) -> Result`: Adds a fragment to the session's rolling window.
    * `tia.SetWindowSize(params)`: Configures the depth and entropy thresholds for aggregation.
* **Data Storage/State:**
    * **Temporal Buffer:** In-memory, high-speed ring buffer for active sessions.
    * **Pattern Registry:** Hardware-attested database of known SIT signatures.

## 5. Alternatives Considered
* **Increasing Fragment-Level Entropy Sensitivity**: Rejected due to high false-positive rates; benign agents often use repetitive patterns.
* **Full Context Window Analysis at Every Step**: Rejected due to prohibitive latency and token costs in 1M+ context windows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The TIA Hub operates within the secure enclave to prevent attackers from "probing" the window limits.
* **Observability:** Aggregated intent patterns and reconstruction events are surfaced in the **ARI Lineage Visualizer**.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation. Introducing Temporal Intent Aggregation to counter Semantic Intent Tunneling (SIT).
