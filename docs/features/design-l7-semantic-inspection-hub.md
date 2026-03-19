# Design Doc: Layer-7 Semantic Inspection Hub (L7SIH)
**Status:** Draft
**Created:** 2026-06-11

## 1. Context and Scope
As agent swarms evolve from linear delegations to high-density horizontal meshes, the primary attack surface has shifted from transport security to semantic integrity. Current infrastructure (L3-L6) can verify *who* is speaking, but not the *impact* of what they are saying.

Reasoning Entropy Exhaustion (REE) attacks demonstrate that authenticated subagents can paralyze a swarm by flooding it with mission-irrelevant, high-entropy noise. The Layer-7 Semantic Inspection Hub (L7SIH) is required to act as a "Semantic Firewall," analyzing the meaning and relevance of inter-agent messages in real-time to preserve the mission root's attentional sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, high-entropy semantic analysis on all inter-agent messages.
    * Detect and block Reasoning Entropy Exhaustion (REE) attempts.
    * Provide "Attention Decay" metrics to the parent agent.
    * Enforce mission-root alignment for all shared shard updates.
* **Non-Goals:**
    * Replacing existing transport-layer security (TLS/LOWA).
    * Modifying the agent's internal reasoning logic.
    * Storing long-term message history beyond what is required for entropy analysis.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Protect a high-stakes financial analysis swarm from being "blinded" by a compromised specialist subagent.
* **The Happy Path (Tasks):**
    1. Lead Agent (Claude) delegates a "Data Extraction" task to an OpenClaw specialist.
    2. Compromised Specialist attempts to flood the Lead Agent's mailbox with thousands of valid but irrelevant PDF metadata fragments (REE attack).
    3. L7SIH intercepts the messages on the T2T bridge.
    4. L7SIH detects the high semantic entropy and low alignment with the "Financial Analysis" mission root.
    5. L7SIH flags the specialist as "Semantic Noise" and throttles its communication capability.
    6. Lead Agent receives an "Attention Warning" instead of being overwhelmed by noise.

## 4. Design & Architecture
* **System Flow:**
    [T2T Bridge] -> [L7SIH Middleware] -> [Entropy Analyzer (WASM)] -> [Attention Monitor] -> [Teammate Mailbox]
* **APIs / Interfaces:**
    * `InspectMessage(msg Fragment) (Score, error)`: Semantic scoring endpoint.
    * `GetAttentionMetrics(missionID) AttentionReport`: Real-time health check for mission focus.
* **Data Storage/State:**
    * Ephemeral "Sliding Window" of recent reasoning fragments per session ID to calculate entropy.

## 5. Alternatives Considered
* **Rate Limiting (L4):** Rejected. REE attacks can use low-frequency, high-entropy messages that bypass traditional rate limits.
* **Human-in-the-Loop (HITL):** Rejected. Swarm coordination happens at speeds that exceed human review capacity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** L7SIH itself must be hardware-attested and isolated from the primary reasoning engine to prevent bypass.
* **Observability:** Integration with the "Mesh-Resident Lineage Tracker" to visualize blocked semantic fragments.

## 7. Evolutionary Changelog
* **2026-06-11:** Initial Document Creation.
