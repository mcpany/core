# Design Doc: Inter-Agent Intent Tracking System

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
In multi-agent swarms, an "Orchestrator" agent often delegates tasks to "Specialist" agents. A critical security and reliability gap exists when the Specialist agent performs actions that deviate from the Orchestrator's original intent, or when a malicious "Semantic Payload" is injected during delegation. The Inter-Agent Intent Tracking System aims to provide a verifiable lineage of intent across these handoffs to prevent cascading failures and unauthorized actions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Maintain a cryptographically signed "Intent Manifest" that follows a task across agent boundaries.
    *   Provide a mechanism for Specialist agents to verify that their delegated task is within the scope of the original parent intent.
    *   Enable the Policy Firewall to block actions that deviate from the signed intent.
    *   Integrate with the A2A Interop Bridge for cross-framework intent propagation.
*   **Non-Goals:**
    *   Automatically correcting agent reasoning (this is a security and observability layer, not a reasoning layer).
    *   Replacing existing authentication mechanisms (it complements them with semantic scoping).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Security Architect.
*   **Primary Goal:** Ensure that a "Customer Support Agent" cannot be tricked into delegating a "Delete Database" task to a "DB Admin Agent" via a semantic exploit.
*   **The Happy Path (Tasks):**
    1.  The parent agent (Orchestrator) creates a task and signs an "Intent Manifest" specifying the allowed scopes (e.g., `scope: read_only, target: customer_records`).
    2.  The Orchestrator delegates the task to a Specialist agent via MCP Any.
    3.  The Specialist agent attempts to call a "Delete" tool.
    4.  MCP Any's Policy Firewall intercepts the call, checks the signed Intent Manifest, and sees that `Delete` is not in the allowed scope.
    5.  The call is blocked, and a "Semantic Deviation" alert is logged.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Creation**: Middleware intercepts the initial parent request and prompts the parent (or uses a side-car LLM) to generate a declarative Intent Manifest.
    - **Propagation**: The manifest is injected into the `Recursive Context Protocol` headers.
    - **Verification**: The `IntentVerificationMiddleware` validates the manifest signature and compares the current tool call against the manifest's allowed scopes.
*   **APIs / Interfaces:**
    - `intent/verify`: Internal API for middleware to check a tool call against a manifest.
    - `intent/sign`: Service for agents to attest to their current intent.
*   **Data Storage/State:** Intent Manifests are stored in the `Shared KV Store` (Blackboard) for the duration of the session.

## 5. Alternatives Considered
*   **Static Policy Rules**: Using fixed Rego policies. *Rejected* because they are too rigid for dynamic agent behaviors; intent must be session-specific.
*   **Manual HITL**: Requiring a human to approve every delegation. *Rejected* because it doesn't scale for autonomous swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Intent Manifest acts as a dynamic capability token. If the manifest is missing or the signature is invalid, all downstream tool calls are restricted to a "Minimal Safe Subset."
*   **Observability:** The "Agent Chain Tracer" in the UI will highlight where an agent attempted to deviate from the parent intent.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
