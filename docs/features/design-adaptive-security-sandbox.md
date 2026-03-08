# Design Doc: Adaptive Security Sandbox

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
With the 255% surge in agentic AI vulnerabilities, a new class of "Adaptive Runtime Exploits" has emerged. Unlike static attacks, these autonomous agents can re-engineer payloads in milliseconds after a failed attempt. Standard, static firewall rules are insufficient. MCP Any requires an **Adaptive Security Sandbox** that can detect high-frequency "trial-and-error" patterns and throttle or block agents exhibiting rogue behavior.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement **Heuristic Rate-Throttling**: Detect and slow down tool-call bursts that suggest automated brute-forcing or adaptive exploitation.
    *   Implement **Intent-Drift Detection**: Monitor if an agent's sequence of tool calls starts to deviate significantly from its initial stated intent.
    *   Introduce **Sub-Network Isolation**: Dynamically restrict the network and filesystem visibility of a tool based on the current session's "Intent-Scope."
*   **Non-Goals:**
    *   Perfectly predicting every possible exploit (focus is on behavioral mitigation).
    *   Replacing the static Policy Firewall (this is an additive layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Protect sensitive internal APIs from an autonomous agent that has been compromised or is behaving erratically.
*   **The Happy Path (Tasks):**
    1.  An agent attempts to call a "ReadDatabase" tool with various SQL injection payloads in rapid succession.
    2.  The Adaptive Sandbox detects the high frequency of failed/rejected calls.
    3.  The Sandbox automatically increases the latency for that specific agent session (Throttling).
    4.  If the behavior continues, the Sandbox triggers a "Hard Block" and alerts the user via the HITL (Human-in-the-Loop) middleware.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception Layer**: Sits within the Middleware pipeline, specifically after the Policy Firewall.
    - **Telemetry Aggregator**: Tracks success/failure rates, call frequency, and payload entropy per session.
    - **Heuristic Engine**: Applies "Leaky Bucket" algorithms for throttling and cosine similarity checks for intent-drift.
*   **APIs / Interfaces:**
    - `POST /v1/security/session/{id}/throttle`: Manually trigger throttling.
    - `GET /v1/security/alerts`: Stream of detected anomalies.
*   **Data Storage/State:** In-memory "Blacklist" and "Scoreboard" for active sessions, with persistence to SQLite for long-term pattern analysis.

## 5. Alternatives Considered
*   **Strict Rate Limiting**: Simple fixed limits (e.g., 5 calls/min). *Rejected* because legitimate complex workflows might exceed these, while adaptive exploits might stay just below them.
*   **LLM-based Monitoring**: Using a second LLM to watch the first. *Rejected* due to latency and cost overhead.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature directly addresses the "Autonomous Reconnaissance" threat.
*   **Observability:** Integrated with the UI's "Connectivity & Security Dashboard" to show real-time "Threat Levels."

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
