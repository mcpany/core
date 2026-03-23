# Design Doc: Reason-Graph Integrity (RGI) Provider
**Status:** Draft
**Created:** [2026-06-18]

## 1. Context and Scope
With the emergence of "Reason-Graph Collision" (RGC) exploits in multi-agent environments, MCP Any requires a mechanism to validate the structural integrity of the reasoning graph before subagent refinements are merged into the mission-root. RGC attacks cause cognitive deadlock by injecting semantically valid but structurally conflicting reasoning nodes.

This document defines the architecture for the RGI Provider, a service that performs hardware-attested graph analysis to ensure mission-root stability.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect circular or conflicting reasoning structures (RGC) in real-time.
    * Provide hardware-attested (TPM) graph validation signatures.
    * Maintain sub-10ms validation latency for reasoning refinement loops.
* **Non-Goals:**
    * Validating the semantic "truthfulness" of reasoning (handled by separate Truth Providers).
    * Governing individual tool-call permissions (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Agent Swarm Orchestrator
* **Primary Goal:** Merge refinements from 5 specialist subagents without triggering a cognitive deadlock.
* **The Happy Path (Tasks):**
    1. Orchestrator receives a "Reasoning Refinement Fragment" from a specialist.
    2. Orchestrator submits the current Mission Reason-Graph and the proposed fragment to the RGI Provider.
    3. RGI Provider performs structural collision analysis.
    4. RGI Provider issues a TPM-signed "Integrity Receipt."
    5. Orchestrator merges the fragment into the mission-root context window.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Refinement] --> B[RGI Provider]
        B --> C{Collision Analysis}
        C -->|Conflict| D[Quarantine Fragment]
        C -->|Safe| E[TPM Signing]
        E --> F[Integrity Receipt]
        F --> G[Mission-Root Merge]
    ```
* **APIs / Interfaces:**
    * `POST /v1/rgi/validate`: Accepts a JSON Reason-Graph object and returns an Integrity Receipt.
    * `GET /v1/rgi/status`: Returns current hardware attestation status.
* **Data Storage/State:**
    * Temporary state-tags for active mission graphs are stored in the hardware-locked memory segment of the RGI Provider.

## 5. Alternatives Considered
* **Recursive LLM Validation:** Rejected due to high latency (1s+) and susceptibility to the same "Attention-Baiting" exploits it should detect.
* **Static Rulesets:** Rejected as Reason-Graphs are too dynamic and non-deterministic for fixed rules.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RGI Provider runs in a hardened enclave with zero access to external network resources.
* **Observability:** Structural collision events are logged with full graph snapshots to the local Security Audit Log.

## 7. Evolutionary Changelog
* **[2026-06-18]:** Initial Document Creation.
