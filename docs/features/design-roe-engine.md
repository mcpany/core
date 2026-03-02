# Design Doc: A2A "Rules of Engagement" (RoE) Engine

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As the agentic ecosystem shifts from centralized orchestration to decentralized meshes (A2A), the risk of unauthorized or unsafe agent-to-agent delegation increases. "Swarm Hijacking" and "Cross-Agent Poisoning" are emerging as major threats. The RoE Engine provides a machine-readable governance framework (aligned with IEEE P3119) to define, enforce, and audit the interaction logic between autonomous agents within the MCP Any gateway.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable administrators to define granular "Rules of Engagement" (e.g., "Agent A can only delegate 'research' tasks to Agent B if the session is marked as 'internal'").
    *   Implement real-time enforcement of RoE for all A2A messages passing through the gateway.
    *   Provide a standardized RoE format that can be exported/imported across MCP Any instances.
    *   Audit all RoE evaluations for compliance reporting.
*   **Non-Goals:**
    *   Defining the specific logic for every possible agent framework (use a standardized translation layer).
    *   Replacing the general-purpose Policy Firewall (RoE is specifically for agent interactions).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Governance Officer.
*   **Primary Goal:** Prevent a customer-facing support agent from delegating tasks to a high-privilege administrative subagent.
*   **The Happy Path (Tasks):**
    1.  Officer defines an RoE in the MCP Any UI: `BLOCK delegation FROM 'support-agent' TO 'admin-agent'`.
    2.  A support agent attempts an A2A call to the admin agent.
    3.  The RoE Engine intercepts the message, evaluates it against the rule, and blocks the request.
    4.  An alert is generated in the Governance Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    `A2A Message -> RoE Engine Middleware -> [Rule Evaluation] -> (Allow/Deny) -> Target Agent`
*   **APIs / Interfaces:**
    - `POST /governance/roe`: Update the active Rules of Engagement.
    - `GET /governance/audit/roe`: Retrieve RoE violation logs.
*   - **RoE Language**: A subset of CEL (Common Expression Language) or Rego optimized for agentic state (sender, receiver, task_type, session_context).
*   **Data Storage/State:** Rules are stored in the internal SQLite DB and cached in memory for high-performance evaluation.

## 5. Alternatives Considered
*   **Hard-coded logic in Adapters**: Too brittle and hard to audit.
*   **Using the general Policy Firewall**: Possible, but RoE needs specific primitives (e.g., `delegation_depth`, `intent_alignment`) that are better handled by a specialized engine.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The RoE Engine is a primary defense against lateral movement in agent swarms.
*   **Observability:** Each A2A interaction in the "Agent Chain Tracer" will be tagged with its RoE evaluation result (e.g., `ROE: PASS [Rule 42]`).

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
