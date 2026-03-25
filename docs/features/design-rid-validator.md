# Design Doc: Recursive Intent Delegation (RID) Validator
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
As AI agent swarms grow in depth, the risk of subagent coercion and permission escalation increases. Recursive Intent Delegation (RID), introduced in UACO v1.8, provides a cryptographic mechanism for parent agents to constrain subagent autonomy. The RID Validator is needed to enforce these constraints at the infrastructure layer, ensuring that subagents cannot mutate intents beyond parent-authorized boundaries or delegate beyond a fixed depth.

## 2. Goals & Non-Goals
* **Goals:**
    * Parse and validate UACO v1.8 RID tokens.
    * Enforce `delegation_depth` limits to prevent infinite subagent spawning.
    * Validate `mutation_boundaries` to ensure subagents stay within their assigned task scope.
    * Implement hash-chaining of intents to provide a non-repudiable audit trail of intent lineage.
* **Non-Goals:**
    * Defining the business logic for subagent task assignment.
    * Replacing the base UACO coordination layer.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a specialized "Researcher" subagent from spawning its own "System Admin" subagent to bypass restricted tool access.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a task to a Researcher with `delegation_depth: 1`.
    2. The Researcher attempts to spawn a third agent with high-trust tool access.
    3. The RID Validator intercepts the request, detects that the depth limit has been reached, and rejects the delegation.
    4. The Researcher is forced to complete the task within its own assigned capabilities.

## 4. Design & Architecture
* **System Flow:**
    `[Delegation Request] -> [RID Validator] -> [Depth Check] -> [Mutation Boundary Check] -> [Audit Logging] -> [Approval/Rejection]`
* **APIs / Interfaces:**
    * Integrated into the UACO Coordinator middleware.
* **Data Storage/State:**
    * Intent lineage is stored in the Shared KV Store (Blackboard) with cryptographic pinning.

## 5. Alternatives Considered
* **Static Permission Lists**: Rejected as they are too rigid for dynamic swarms and don't account for delegation lineage.
* **Simple Session Tokens**: Rejected as they lack depth and mutation awareness.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RID is a core component of the Zero-Trust mesh, moving from "Identity-based" to "Intent-based" security.
* **Observability:** Delegation depth and boundary violations are alerted in the "Recursive Loop Heatmap."

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
