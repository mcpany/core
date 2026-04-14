# Design Doc: Adaptive Node Quarantining (ANQ) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms evolve toward multi-node Sovereign Node Tunneling (SNT), the risk of "Cognitive Hijacking" increases. A single compromised or drifting node can pollute the entire mission-root blackboard. MCP Any needs a mechanism to monitor the semantic health of remote nodes and isolate them autonomously.

The ANQ Hub leverages the Agentic Entropy Monitor (AEM) to perform real-time analysis of subagent reasoning. When entropy scores exceed safe thresholds, the hub interdicts the node's communication tunnels, preventing further state mutation.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide real-time reasoning entropy scoring for all remote mesh nodes.
    * Automatically revoke SNT tunnel access upon threshold violation.
    * Maintain a "Quarantine Log" for forensic analysis of drifted reasoning paths.
* **Non-Goals:**
    * Automatically "fixing" the drifted reasoning (requires human or supervisor intervention).
    * Monitoring nodes not connected via the Universal Agent Bus.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a drifting specialist node from executing unauthorized filesystem edits on the host.
* **The Happy Path (Tasks):**
    1. The user defines an ANQ policy with a maximum entropy threshold of 0.85 for the "Filesystem Specialist" role.
    2. A remote node is delegated a task and begins reasoning.
    3. The node is hit with a prompt injection attempt hidden in a retrieved file.
    4. The node's reasoning entropy spikes to 0.92 as it attempts to rationalize the injection.
    5. The ANQ Hub detects the spike via the AEM integration.
    6. The Hub immediately terminates the node's SNT tunnel and locks its Blackboard shard.
    7. The user is notified via the Swarm Health Dashboard.

## 4. Design & Architecture
* **System Flow:**
    [Node Reason Fragment] -> [AEM Scoring] -> [ANQ Evaluator] -> [Tunnel Controller (Kill-Switch)]
* **APIs / Interfaces:**
    * `POST /v1/mesh/quarantine/{node_id}`: Manual override to isolate a node.
    * `GET /v1/mesh/health`: Returns real-time entropy scores and quarantine status for all nodes.
* **Data Storage/State:**
    * State is managed via the `Resilience Broker`, which persists node status (Active/Quarantined) and reasoning-path fingerprints.

## 5. Alternatives Considered
* **Manual Revocation Only:** Rejected due to the machine-speed nature of "Hivenet" attacks; human-in-the-loop is too slow for containment.
* **Global Swarm Lockdown:** Rejected as it impacts mission availability; granular node-level isolation is preferred.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The quarantine signal itself must be hardware-attested to prevent subagents from "Quarantining" their supervisors.
* **Observability:** Real-time entropy heatmaps are streamed to the Swarm Health Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
