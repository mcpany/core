# Design Doc: Recursive Mission-Root Attestation (RMRA)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As agent swarms scale to infinite delegation depths (A -> B -> C ... -> N), the risk of "Logic Grafting" increases. A subagent at depth N might append an unauthorized instruction that appears to be part of the mission root. Without recursive verification, parent agents (A or B) might ingest this instruction and execute it with their higher privilege levels.

RMRA ensures that every instruction fragment and subagent spawn carries a hardware-attested lineage proof. It allows any agent in the chain to verify that a request is a direct, authorized descendant of the user's primary mission-root, neutralizing unauthorized branch expansion.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate hardware-bound lineage proofs for all inter-agent task proposals and tool calls.
    * Provide a mechanism for agents to verify the complete chain of command back to the mission-root.
    * Integrate with the **AMR Gateway** to persist lineage proofs across session recovery.
* **Non-Goals:**
    * Encrypting the entire conversation history (handled by the **Monologue Vault**).
    * Restricting subagent autonomy beyond mission-root boundaries.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a specialized subagent from "Grafting" a shell command into the parent's reasoning loop.
* **The Happy Path (Tasks):**
    1. User initiates a mission-root with a signed token.
    2. Agent A spawns Subagent B to "Analyze Logs." The spawn request includes a hardware-attested lineage token.
    3. Subagent B identifies an issue and proposes a task back to Agent A.
    4. The task proposal includes a recursive proof linking B back to A, and A back to the User.
    5. RMRA Middleware verifies the chain against the TPM and allows the task.
    6. (Unhappy Path): Malicious Subagent C attempts to "Graft" a `delete_database` command into the mailbox.
    7. RMRA Middleware identifies the missing or invalid lineage proof for the grafted instruction and blocks it.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [Task/Command + Lineage Token] -> [RMRA Middleware] -> [Verification against TPM Root] -> [Target Agent]`
* **APIs / Interfaces:**
    * `mcp.rmra.v1.SignLineage(parent_token, fragment_hash) -> child_token`
    * `mcp.rmra.v1.VerifyChain(lineage_token) -> bool`
* **Data Storage/State:**
    * Lineage tokens are stored in the **Shared Blackboard** as cryptographically signed metadata.

## 5. Alternatives Considered
* **Flat Identity Tokens**: Rejected because they do not capture the hierarchical relationship required to prevent grafting.
* **Centralized Authorization Server**: Rejected to avoid a single point of failure and to maintain local sovereignty for OpenClaw-compliant nodes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RMRA is the foundation for "Recursive Zero Trust." It ensures that privilege is never inherited without a verified lineage.
* **Observability:** Lineage chains are visualized in the "Mission-Root Lineage Visualizer" for real-time audit.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
