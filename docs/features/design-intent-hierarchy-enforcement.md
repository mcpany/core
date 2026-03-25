# Design Doc: Intent Hierarchy Enforcement

**Status:** Draft

**Created:** 2026-05-30

## 1. Context and Scope

Swarms frequently experience "Context Shadowing," where subagents override root
mission constraints. This feature introduces an authoritative hierarchy.

## 2. Goals & Non-Goals

- **Goals:** Enforce intent priority based on lineage.
- **Non-Goals:** It will not modify the underlying LLM weights.

## 3. Critical User Journey (CUJ)

- **User Persona:** Swarm Orchestrator
- **Primary Goal:** Prevent subagent from overriding mission-root
- **Steps:**
  1. User sets mission-root.
  2. Subagent attempts override.
  3. Middleware blocks override.

## 4. Design & Architecture

- **System Flow:** Middleware intercepts tool calls.
- **APIs:** `POST /api/v1/hierarchy/validate`

## 5. Alternatives Considered

- Regex (rejected: too brittle).

## 6. Cross-Cutting Concerns

- **Security:** Neutralizes semantic injection.

## 7. Evolutionary Changelog

- **2026-05-30:** Initial Document Creation.
