# Design Doc: Swarm Anomaly Detection Engine
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
With the rise of "Hivenet" swarm attacks, traditional single-call security policies (like the Policy Firewall) are insufficient. Attackers now distribute malicious tasks across multiple autonomous agents (e.g., OpenClaw, Claude Code) that individually appear benign but collectively execute a high-impact breach. MCP Any, as the central gateway for these swarms, is uniquely positioned to perform cross-session anomaly detection.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect coordinated probing or exploitation patterns across multiple agent sessions.
    * Provide real-time "Swarm Risk" scoring for tool calls.
    * Automatically throttle or suspend suspicious agent groups.
    * Correlate tool usage across different agent frameworks (A2A).
* **Non-Goals:**
    * Solving individual prompt injection (handled by Policy Firewall).
    * General network IDS/IPS (focus is specifically on Agentic Tool Call patterns).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Operations Center (SOC) Analyst.
* **Primary Goal:** Identify and block a coordinated attempt by a swarm to exfiltrate database records via multiple small, distributed queries.
* **The Happy Path (Tasks):**
    1. A "Hivenet" swarm of 10 subagents is spawned via an indirect prompt injection.
    2. Each subagent calls `sql_query` with a small `LIMIT 10` on different tables.
    3. The Anomaly Detection Engine aggregates these calls in the "Global Blackboard."
    4. Pattern matching identifies a "Distributed Scanning" signature.
    5. The Engine raises the risk score for the parent session.
    6. The HITL Middleware is triggered, suspending all 10 subagents and alerting the SOC.

## 4. Design & Architecture
* **System Flow:**
    - **Pattern Ingestion**: Tool calls from all active sessions are streamed to an in-memory "Sliding Window" buffer.
    - **Correlation Engine**: Matches sequences of calls against known "Swarm Attack Signatures" (e.g., port scanning, credential stuffing, distributed exfiltration).
    - **Risk Scoring**: Calculates a dynamic `SwarmConfidence` score.
* **APIs / Interfaces:**
    - `POST /v1/anomaly/signals`: Ingest external signals from agent frameworks.
    - `GET /v1/anomaly/swarm-status`: Current active swarm risk levels.
* **Data Storage/State:** High-performance Redis or in-memory TTL-based store for the sliding window of events.

## 5. Alternatives Considered
* **Agent-Side Detection**: Embedding detection in each agent. *Rejected* because agents cannot see each other's activity, making it impossible to detect distributed swarms.
* **SIEM-Based Detection**: Sending all logs to a corporate SIEM. *Rejected* because the latency of SIEM processing is too high for real-time agent suspension.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The engine itself must be protected from "Poisoning" where an agent tries to trigger false positives for other agents.
* **Observability:** Integration with the "Agent Chain Tracer" to visualize the swarm's coordinated actions.

## 7. Evolutionary Changelog
* **2026-02-26:** Initial Document Creation.
