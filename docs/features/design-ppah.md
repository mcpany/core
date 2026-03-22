# Design Doc: Privacy-Preserving A2A Handoffs (PPAH)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
The "InversePrompt" (CVE-2025-54795) vulnerability demonstrated that LLM agents can be turned against their own security restrictions via recursive self-interpretation. When an agent delegates a task to another, it often shares excessive context, which can include sensitive system prompts or organizational constraints. This "Inherited Context" becomes an attack vector for prompt injection in the recipient agent.

PPAH provides a security layer for A2A handoffs that enforces **Context Minimization**. Instead of sending the full conversation history or system prompt, PPAH ensures that only a minimized "Intent Fragment" is shared. This limits the "Semantic Surface Area" available for recursive exploits and protects organizational secrets during inter-agent collaboration.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time semantic pruning for all A2A task delegation messages.
    * Ensure that "Intent Fragments" are consistent with the mission-root but stripped of system-level metadata.
    * Neutralize "Inherited Context Poisoning" by enforcing Zero-Trust at the handoff boundary.
* **Non-Goals:**
    * Encrypting the payload (handled by the T2T Encryption Bridge).
    * Modifying the recipient agent's internal reasoning logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Delegate a "Code Fix" task to a specialized subagent without exposing the parent agent's private system-prompt instructions regarding security guardrails.
* **The Happy Path (Tasks):**
    1. The parent agent initiates an A2A task proposal for a "Code Fix."
    2. PPAH intercepts the proposal before it is sent to the messaging hub.
    3. PPAH performs semantic analysis, identifying fragments of the parent's system prompt.
    4. PPAH prunes the context, leaving only the task description and relevant code snippets.
    5. The minimized "Intent Fragment" is delivered to the subagent.
    6. The subagent performs the fix without exposure to the parent's internal constraints.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] -> (Task Proposal) -> [PPAH Middleware] -> (Minimization/Pruning) -> [A2A Messaging Hub] -> [Subagent]
* **APIs / Interfaces:**
    * `PPAH.Minimize(proposal TaskProposal) TaskProposal`: The core minimization engine.
    * `PPAH.Validate(fragment IntentFragment)`: Checks if a fragment is "Safe-for-Handoff."
* **Data Storage/State:**
    * Pruning rules are managed by the Policy Engine and can be customized per profile.

## 5. Alternatives Considered
* **Manual Sanitization:** Rejected because agents cannot be trusted to self-sanitize their own context when compromised.
* **Fixed Context Windows:** Rejected because fixed limits often strip necessary data while leaving system-prompt fragments intact. Semantic minimization is required.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PPAH acts as a mandatory egress filter for all inter-agent communication.
* **Observability:** Pruning events and minimization ratios are visualized in the "Privacy Handoff Status" dashboard.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
