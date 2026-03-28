# Design Doc: Action-Chain Sovereignty Monitor (ACSM)
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
Autonomous agents operate at speeds that exceed human intervention capabilities. A single compromised intent can trigger a cascade of cross-system actions (e.g., file edit -> commit -> build -> deploy), leading to massive system failure. ACSM provides real-time monitoring and validation of complete workflow sequences to ensure they remain anchored to the verified mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Track and validate the entire sequence of actions initiated by an agent.
    * Provide sub-millisecond interdiction of action chains that diverge from mission-root intent.
    * Maintain a cryptographically signed "Chain of Command" for every automated workflow.
* **Non-Goals:**
    * Providing general-purpose application logging (handled by standard observability stacks).
    * Predicting future agent actions (ACSM is reactive/validating).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Center (SOC) Analyst
* **Primary Goal:** Detect and block an agent attempting to exfiltrate database secrets through a series of "innocent" code analysis steps.
* **The Happy Path (Tasks):**
    1. Agent begins a "Security Audit" mission.
    2. ACSM initializes a new "Action Chain" token bound to this mission.
    3. Agent performs a series of file reads (authorized).
    4. Agent attempts to call a network egress tool with sensitive data fragments.
    5. ACSM identifies the drift in the action-chain logic and revokes the agent's egress capability instantly.

## 4. Design & Architecture
* **System Flow:**
    `Action Request` -> `ACSM Validator` -> `Mission-Root Oracle` -> `Gatekeeper (Allow/Deny)`
* **APIs / Interfaces:**
    * `ActionChainMonitor`: `RegisterAction(actionID string, payload []byte) error`, `ValidateChain(sessionID string) (bool, error)`
* **Data Storage/State:**
    * Action chains are stored as a hash-chained sequence in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Per-Action HITL**: Rejected as it defeats the purpose of autonomous agents and leads to approval fatigue.
* **Static RBAC**: Rejected because it cannot detect malicious logic in a sequence of otherwise "safe" actions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ACSM uses "Semantic Entropy" analysis to detect low-and-slow probes that might bypass simple intent-matching.
* **Observability:** Action-chain interdictions are visualized in the "Multi-Agent Swarm Topology Monitor."

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
* **2026-03-28:** Update: Mandatory Intent-Bound Action Mapping.
    **Context:** Today's market sync revealed coordinated "Chain-of-Failure" exploits in autonomous triage swarms.
    **Architecture Adjustment:**
    * Transitioning from passive chain logging to active "Intent-Bound Action Mapping" in Section 4.
    * Integrating real-time "Semantic Entropy" analysis for multi-step drift detection.
    **Security Impact:** Prevents "Death by 1,000 Cuts" exfiltration patterns by revoking capabilities upon the first sign of action-chain divergence.
