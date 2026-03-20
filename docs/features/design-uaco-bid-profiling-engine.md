# Design Doc: UACO Bid Profiling Engine
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
The Universal Agent Coordination Protocol (UACO) provides a mechanism for agent task bidding. However, research into "Task Card Shadowing" (CVE-2026-34015) and "Negotiation Storms" reveals that a compromised agent can attempt to "shadow" or bid on unauthorized tasks to gain elevated access. MCP Any needs an "Active Behavioral Validator" that performs "Pre-Flight Profiling" on agents submitting UACO bids. This ensures that only agents with a verified historical safety profile and relevant "Skill Profile" are eligible to bid.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Pre-Flight Profiling" for all UACO task bids.
    * Compare the bidding agent's current "Skill Profile" and "Trust Level" against the task card requirements.
    * Automatically quarantine bids from agents with anomalous historical behavior.
    * Provide a "Safety Score" for every bid to inform parent-agent selection logic.
* **Non-Goals:**
    * Fully automating the selection process (remains the responsibility of the parent agent).
    * Restricting *all* bids (only those with high-risk mismatches).

## 3. Critical User Journey (CUJ)
* **User Persona:** Parent Agent Orchestrator
* **Primary Goal:** Select a "Database Specialist" subagent from a pool of bidders to perform a migration task.
* **The Happy Path (Tasks):**
    1. The parent agent broadcasts a `UACO.taskCard` for "DB Migration."
    2. Multiple subagents submit `UACO.bid` objects.
    3. The Bid Profiling Engine intercepts each bid.
    4. For Agent A, it verifies that their skill registry includes "Database Admin" and their historical trust level is "P0."
    5. For Agent B (compromised), it detects that their recent behavior includes "Unauthenticated Port Scans."
    6. Agent B's bid is automatically quarantined.
    7. Agent A's bid is returned to the parent with a high "Safety Score."

## 4. Design & Architecture
* **System Flow:**
    `UACO.bid` -> `Bid Profiling Engine` -> `Reputation Engine` -> `Safety Scorer` -> `Parent Agent`
* **APIs / Interfaces:**
    * `UACO.getSafetyScore(bid_id)`
    * `UACO.quarantineBid(bid_id, reason)`
* **Data Storage/State:**
    * Historical behavior is stored in the "Unified Telemetry Bridge."

## 5. Alternatives Considered
* **Static Bidding Rules**: Rejected because they are too easily bypassed by "Logic Grafting" and "Stylometric Mimicry."
* **Manual Review for all bids**: Rejected as a major bottleneck for autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The profiling logic must be hardware-bound to prevent a compromised agent from "spoofing" their own safety score.
* **Observability:** Quarantined bids and low safety scores are logged to the "Enterprise Policy Sync Engine."

## 7. Evolutionary Changelog
* **2026-03-20:** Initial Document Creation.
