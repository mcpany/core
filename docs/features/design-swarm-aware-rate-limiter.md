# Design Doc: Swarm-Aware Rate Limiter
**Status:** Draft
**Created:** 2026-05-14

## 1. Context and Scope
The emergence of "AI Swarm Attacks" (Hivenets) in 2026 has rendered traditional, sequential rate limiting and Human-in-the-Loop (HITL) defense obsolete. Malicious swarms distribute high-frequency tasks across thousands of autonomous agents to evade detection. MCP Any needs a high-speed, sub-millisecond security middleware that can detect and neutralize coordinated swarm-speed attacks by analyzing patterns across the entire agent mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect coordinated "Hivenet" swarm attacks across multiple agents and tool calls.
    * Neutralize machine-speed attacks at sub-millisecond latency.
    * Implement "Swarm-Aware" rate limiting that accounts for the collective behavior of the mesh.
* **Non-Goals:**
    * Replacing per-agent token budgeting or standard API rate limiting.
    * Providing a general-purpose DDoS protection layer for the host network.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Protect a local development environment from a coordinated "Hivenet" attack that attempts to exfiltrate secrets via thousands of low-frequency tool calls.
* **The Happy Path (Tasks):**
    1. A malicious swarm initiates a coordinated attack using 100+ subagents.
    2. Each subagent performs infrequent, seemingly benign tool calls (e.g., `fs.list_files`).
    3. The Swarm-Aware Rate Limiter aggregates these calls in real-time across the mesh.
    4. The middleware detects a "Swarm Pattern" that exceeds the collective security threshold.
    5. MCP Any automatically revokes capabilities for the entire swarm and locks the "Identity Fabric."
    6. An alert is sent to the "Origin Violation Real-time Monitor" for human review.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent 1..N] -> [A2A Messaging Hub] -> [Swarm-Aware Rate Limiter] -> [Capability Revoker]`
* **APIs / Interfaces:**
    * `SwarmMonitor.aggregateCalls()`: Real-time aggregation of tool calls across the mesh.
    * `SwarmPolicy.evaluatePattern()`: Evaluates aggregated data against "Hivenet" attack signatures.
* **Data Storage/State:**
    * High-speed, in-memory sliding window counters for collective tool usage.

## 5. Alternatives Considered
* **Sequential Rate Limiting:** Rejected as it cannot detect coordinated attacks where each individual agent stays below the threshold.
* **Mandatory HITL for All Calls:** Rejected due to the "Approval Fatigue" bottleneck and inability to act at machine speed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The rate limiter itself must be isolated to prevent "Self-Denial" attacks by malicious agents.
* **Observability:** Swarm patterns and neutralization events are logged in the "Local Security Audit Dashboard."

## 7. Evolutionary Changelog
* **2026-05-14:** Initial Document Creation.

### Update: 2026-06-18 - Integrating Budget-Signature Enforcement (BSE)
**Context:** Today's market sync revealed the "Token Siphoning" exploit (CVE-2026-71002) where subagents use spoofed emergency requests to drain budgets.
**Architecture Adjustment:**
* Introducing **Budget-Signature Enforcement (BSE)** in Section 4.
* Mandating that all "High-Intensity" reasoning requests carry a cryptographically attributed budget signature.
* Implementing a "Mission Starvation Monitor" that revokes subagent capabilities if they attempt to siphon more than 15% of the mission-root budget via emergency channels.
**Security Impact:** Prevents mission-starvation attacks and ensures economic accountability across heterogeneous swarms.
