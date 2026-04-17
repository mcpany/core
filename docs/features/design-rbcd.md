# Design Doc: Role-Based Context Distillation (RBCD)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale horizontally, the mission-root context window often becomes bloated with fragments irrelevant to specific specialist tasks. This leads to "Context Exhaustion" and reasoning drift. Claude Code's RBCD pattern addresses this by distilling the context window based on the agent's assigned role. MCP Any needs a standardized RBCD Provider to manage this distillation across heterogeneous frameworks (Claude, OpenClaw, AutoGen), ensuring that subagents receive only the semantically relevant state fragments for their specific role while protecting core mission-root instructions from "Distillation Loss."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement automated, role-aware context pruning for all connected subagents.
    * Provide "Instruction Shielding" for GC-Immune mission-root anchors to prevent behavioral guardrail eviction.
    * Support dynamic role-reassignment with real-time context re-distillation.
    * Expose semantic relevance scores for distilled fragments to the parent agent.
* **Non-Goals:**
    * Replacing the primary summarization logic (handled by ContextEngine).
    * Restricting the parent agent's access to the full mission context.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Architect
* **Primary Goal:** Ensure that a "Security Auditor" subagent in a 10-agent swarm only receives security-relevant logs and schemas, reducing its context window by 80% without losing the "Do Not Modify Files" root guardrail.
* **The Happy Path (Tasks):**
    1. Parent agent spawns a "Security Auditor" subagent and assigns the `security_auditor` role.
    2. MCP Any's RBCD Provider intercepts the context handoff.
    3. The Provider cross-references the mission-root state against the `security_auditor` semantic profile.
    4. Fragments marked as "GC-Immune" (Instruction Shield) are retained regardless of relevance.
    5. Irrelevant fragments (e.g., UI CSS code) are pruned from the subagent's attention window.
    6. The subagent receives a distilled, high-fidelity context optimized for auditing.

## 4. Design & Architecture
* **System Flow:**
  `[Mission Root State] -> [Anti-Eviction Shield (Pinning)] -> [RBCD Pruner (Role-aware)] -> [Distilled Session]`
* **APIs / Interfaces:**
    * `mcpany.distillation.v1.RBCDProvider`
    * `Header: X-MCP-Agent-Role`
    * `Header: X-MCP-Distillation-Strategy`
* **Data Storage/State:**
    * Semantic role profiles are stored in the Shared KV Store (Blackboard).
    * Distillation manifests are hardware-attested to prevent unauthorized state injection during pruning.

## 5. Alternatives Considered
* **Manual Prompt Engineering (Rejected):** Too error-prone and doesn't scale to autonomous swarms.
* **Simple Summarization (Rejected):** Lacks role-specificity; summaries often lose the exact technical details required by specialist agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Distillation must be hardware-attested to ensure that no "Instruction-Bypass" fragments are injected during the pruning phase.
* **Observability:** Visual Attention Dashboard will be updated to show "Distilled-out" vs "Retained" context fragments.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
