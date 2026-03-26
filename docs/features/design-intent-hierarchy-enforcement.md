# Design Doc: Intent Hierarchy Enforcement

**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope

Swarms frequently experience "Context Shadowing," where subagents override root
mission constraints. This feature introduces an authoritative hierarchy that
ensures mission alignment and prevents unauthorized overrides.

## 2. Goals & Non-Goals

- **Goals:**
  - Enforce intent priority based on command lineage.
  - Provide a hierarchical store for mission-root constraints.
- **Non-Goals:**
  - Modifying the underlying LLM weights.

## 3. Critical User Journey (CUJ)

- **User Persona:** Swarm Orchestrator
- **Primary Goal:** Prevent a subagent from overriding a mission-critical
  security constraint.
- **The Happy Path:**
  1. User sets a mission-root constraint via MAH.
  2. Subagent attempts to override the instruction in its local context.
  3. IHE middleware detects the shadowing attempt and blocks the tool call.

## 4. Design & Architecture

- **System Flow:** Middleware intercepts tool calls and validates state
  fragments against the Intent Hierarchy store.
- **APIs:** `POST /api/v1/hierarchy/validate`

## 5. Alternatives Considered

- **Regex-based blocking:** Rejected as too brittle for semantic overrides.

## 6. Cross-Cutting Concerns

- **Security:** Neutralizes semantic injection.
- **Observability:** Detailed audit logs for all blocked override attempts.

## 7. Evolutionary Changelog

- **2026-05-30:** Initial Document Creation.
