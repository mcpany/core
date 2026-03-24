# Design Doc: Cognitive Attestation Hub (CAH) Adapter
**Status:** Draft
**Created:** 2026-06-30

## 1. Context and Scope
With the release of OpenClaw v3.3.0, the industry is moving from point-to-point tool security to "Collective Cognitive Consensus." Swarms now require a way to reach agreement on the integrity of a reasoning path before any side effects are committed to the shared environment.

The **CAH Adapter** enables MCP Any to participate in this consensus model. It provides the standardized hooks and multi-signature orchestration required for swarms to attest to reasoning integrity, neutralizing "Hallucination Variance" and "Ghost Reasoning" in complex meshes.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement OpenClaw-compatible CAH consensus hooks.
    *   Orchestrate hardware-attested approval tokens from multiple monitor agents.
    *   Block Blackboard commits for tasks that fail the "Cognitive Quorum."
    *   Standardize the `X-CAH-Signature` header for all inter-framework coordination.
*   **Non-Goals:**
    *   Performing the actual reasoning (handled by specialist agents).
    *   Replacing the standard Policy Firewall (CAH acts as a pre-commit cognitive gate).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Compliance Officer
*   **Primary Goal:** Ensure that a code-refactoring subagent only commits changes if a "Security Auditor" agent and a "Quality Monitor" agent both reach a hardware-attested consensus on the refactor's safety.
*   **The Happy Path (Tasks):**
    1.  Lead Agent proposes a code refactor task.
    2.  Refactor Agent generates the code change.
    3.  CAH Adapter intercepts the commit request and triggers a "Cognitive Quorum."
    4.  Auditor and Monitor agents review the reasoning trace and provide TPM-signed attestation tokens.
    5.  CAH Adapter aggregates the tokens and verifies the `X-CAH-Signature`.
    6.  The commit is finalized on the Blackboard.

## 4. Design & Architecture
*   **System Flow:**
    `[Agent] -> [CAH Adapter] -> [Quorum Trigger] -> [Signature Aggregator] -> [Blackboard Commit]`
*   **APIs / Interfaces:**
    *   `POST /cah/propose`: Submit a reasoning fragment for consensus.
    *   `GET /cah/quorum-status`: Poll for attestation results.
    *   `X-CAH-Signature`: Multi-sig JWT containing hardware-bound attestations.
*   **Data Storage/State:**
    *   Ephemeral Quorum Registry tracking active consensus attempts and participant heartbeats.

## 5. Alternatives Considered
*   **Sequential HITL**: Rejected as it creates massive bottlenecks in autonomous swarms.
*   **Centralized AI Auditor**: Rejected because it represents a single point of failure; collective consensus provides a stronger security posture.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All participants must provide hardware-bound identity tokens (TPM/Secure Enclave).
*   **Observability:** Integrated with the "Swarm Anomaly Visualizer" to show quorum strengths and participant dissent.

## 7. Evolutionary Changelog
*   **2026-06-30:** Initial Document Creation.
