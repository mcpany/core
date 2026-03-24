# Design Doc: AIR Reconciliation Broker
**Status:** Draft
**Created:** 2026-05-21

## 1. Context and Scope
In decentralized agent swarms (e.g., Claude Code Agent Teams, OpenClaw swarms), conflicting instructions and intents are common. "Autonomous Intent Reconciliation" (AIR) is required to determine the "Winning Intent" without human intervention. MCP Any, as the Universal Agent Bus, must provide the infrastructure to resolve these conflicts using verifiable, hardware-attested quorums.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized service for reconciling conflicting intents across heterogeneous agent frameworks.
    * Utilize hardware-attested multi-signature quorums to provide a verifiable "Winning Intent."
    * Support "Conflict-Free Replicated Reasoning" (CFRR) signals for merging non-conflicting traces.
* **Non-Goals:**
    * Replacing the reasoning engines of individual agents.
    * Enforcing intents that violate the user's root mission.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Resolve a conflict between two subagents regarding a critical filesystem edit.
* **The Happy Path (Tasks):**
    1. Subagent A proposes Intent X (edit file); Subagent B proposes Intent Y (delete file).
    2. MCP Any detects the conflict on the Shared Blackboard.
    3. The AIR Broker triggers an Intent Reconciliation session.
    4. Independent "Auditor" agents provide signatures for the intent that best aligns with the Mission Root.
    5. The AIR Broker publishes the hardware-attested "Winning Intent" to the swarm.

## 4. Design & Architecture
* **System Flow:**
    `[Conflicting Intents] -> [AIR Broker] -> [Consensus Hub] -> [Attested Winning Intent]`
* **APIs / Interfaces:**
    * `AIR.reconcile_intents(intent_list, mission_root_id)`: Initiates reconciliation for a list of intents.
    * `AIR.submit_signature(session_id, agent_identity, signature)`: Allows quorum members to vote.
* **Data Storage/State:**
    * Reconciliation sessions and winning intents are stored in the Versioned State Hub (Blackboard).

## 5. Alternatives Considered
* **Last-Writer-Wins (LWW)**: Rejected as it doesn't account for mission alignment or security risks.
* **Mandatory HITL**: Rejected as it creates a performance bottleneck in deep autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All reconciliation signatures must be cryptographically bound to hardware-attested identities.
* **Observability**: Reconciliation events and winning intent lineage are tracked in the Swarm Truth Explorer.

## 7. Evolutionary Changelog
* **2026-05-21:** Initial Document Creation.
