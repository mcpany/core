# Design Doc: Branch-Purity Blackboard Validator
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
As swarms move from linear execution to parallel branch exploration (e.g., OpenClaw's "Reasoning-Bound Context Shifting"), the Shared KV Store (Blackboard) becomes vulnerable to "Branch Contamination." If an agent explores multiple hypothetical paths, state from a discarded branch can leak into the primary mission context, leading to hallucinations. MCP Any needs a validation layer that enforces "Branch Purity" for all state mutations.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Branch-Bound" namespaces for all Blackboard writes.
    * Enforce a "Parental Re-Attestation" requirement before merging hypothetical results back into the primary intent chain.
    * Automatically purge state from pruned or non-convergent reasoning branches.
* **Non-Goals:**
    * Managing per-agent private memory (which is handled by SRM).
    * Resolving semantic conflicts between branches (handled by MRCR).

## 3. Critical User Journey (CUJ)
* **User Persona:** Parallel Team Orchestrator
* **Primary Goal:** Agent A explores "Plan X" and Agent B explores "Plan Y" in parallel. If Plan X is rejected, Plan X's state must not be visible to Agent B.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns two parallel intent branches (X and Y).
    2. Validator creates isolated **Branch Shards** on the Blackboard.
    3. Agent A writes a fact to Shard X.
    4. Mission Root evaluates both branches and chooses Plan Y.
    5. Validator forcefully purges Shard X.
    6. Shard Y is promoted to the mission-root context after user/supervisor attestation.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Write] -> [Blackboard Validator] -> [Branch Metadata Tagging] -> [Isolated Shard]`
    `[Pruning Signal] -> [Validator Reaper] -> [Shard Deletion]`
* **APIs / Interfaces:**
    * `blackboard/branch/create`: Issues a new isolated branch ID.
    * `blackboard/branch/merge`: Promotes a branch shard to the parent context.
* **Data Storage/State:**
    Blackboard keys are prefixed with `[branch_id]:` and utilize a hierarchical tagging system to track parentage.

## 5. Alternatives Considered
* **Flat Memory with Rollbacks:** Rejected due to performance overhead in high-density swarms. Isolated shards are more scalable.
* **Agent-Private Storage only:** Rejected because teammates need to share state within a valid branch.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Branches are cryptographically sealed to their specific intent-token.
* **Observability:** Logs "Branch Drift" metrics and the ratio of pruned to promoted state fragments.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
