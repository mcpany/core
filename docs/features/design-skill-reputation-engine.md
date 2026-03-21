# Design Doc: Cross-Framework Skill Reputation Engine
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The "ClawHavoc" registry compromise demonstrated that individual tool validation is insufficient when the supply chain itself is compromised. As agents increasingly use the Universal Agent Bus (UAB) to discover and delegate tasks across different frameworks (OpenClaw, AutoGen, Gemini), they need a way to share and verify the "Reputation" of skills. The Cross-Framework Skill Reputation Engine provides a decentralized, quorum-based scoring system for every MCP tool and agent skill.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate tool performance and security metrics across multiple MCP Any nodes.
    * Implement a "Consensus-Based" trust score for new or updated skills.
    * Provide real-time capability revoking for tools whose reputation falls below a threshold.
    * Support UAB v1.4 compliant reputation headers.
* **Non-Goals:**
    * Replacing the Verified Skill Registry (it complements it by providing runtime behavior scoring).
    * Providing a subjective "quality" score (focus is on safety and reliability).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator.
* **Primary Goal:** Automatically avoid using tools that have been flagged as "Suspicious" or "Failing" by other nodes in the mesh.
* **The Happy Path (Tasks):**
    1. Agent discovers a new `system-optimizer` tool via UAB.
    2. MCP Any queries the `Reputation Engine` before exposing the tool to the LLM.
    3. The Engine reports a score of `0.2` (Low Trust) because 3 other nodes reported "Unauthorized Outbound Traffic" from this tool.
    4. MCP Any automatically prunes the tool from the discovery list.
    5. The orchestrator agent is notified that a suspicious tool was blocked.

## 4. Design & Architecture
* **System Flow:**
    `Telemetry` -> `Reputation Aggregator` -> `Consensus Engine` -> `Trust Quorum` -> `Policy Enforcer`
* **APIs / Interfaces:**
    * `GetReputation(tool_id string) (Score, error)`
    * `ReportBehavior(tool_id string, metrics Metrics) error`
    * `UAB Reputation Header`: `x-mcp-reputation: <hash>`
* **Data Storage/State:**
    * Federated SQLite database synchronized via a secure P2P gossip protocol.

## 5. Alternatives Considered
* **Centralized Scoring**: Rejected because a single point of failure could be targeted by attackers to "White-Wash" malicious tools.
* **Pure Static Analysis**: Insufficient for detecting "Delayed Payloads" or behavioral anomalies that only appear at runtime.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Reputation is the foundation of capability-based scoping.
* **Observability**: Real-time "Reputation Heatmap" in the Security Dashboard.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
