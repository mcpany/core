# Design Doc: Plan-Mode Verification Middleware (PVN)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the standardization of "Plan Mode" in agents like Gemini CLI and Claude Code, agents now propose a sequence of actions before execution. However, market research reveals that these plans themselves are becoming attack vectors. "Plan-Level Injection" allows attackers to coerce agents into multi-step malicious workflows (e.g., read secret -> obfuscate -> exfiltrate) that individual tool-call monitors might miss because each step looks benign in isolation.

MCP Any needs to solve this by providing a unified middleware that audits the **entire plan** against mission-root constraints before the first tool is ever invoked. PVN acts as the final gatekeeper for autonomous sequences.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate multi-step agent plans before execution.
    * Provide hardware-attested approval tokens for validated plans.
    * Detect "Plan-Level Injection" patterns (e.g., data flow from sensitive sources to external sinks).
    * Support cross-framework plan formats (Gemini, Claude, OpenClaw).
* **Non-Goals:**
    * Replacing existing tool-call level security (PVN is an additional layer).
    * Automatically fixing malicious plans (it will block and alert instead).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Ensure that a terminal agent doesn't execute a sequence that exfiltrates local `.env` files even if "Plan Mode" is enabled.
* **The Happy Path (Tasks):**
    1. Agent generates a 5-step plan to "Fix a bug in the auth module."
    2. Agent sends the plan to the MCP Any PVN endpoint.
    3. PVN performs semantic analysis, identifying that step 2 reads `.env` and step 5 sends a webhook.
    4. PVN flags the plan as "High Risk: Potential Exfiltration."
    5. MCP Any blocks the agent and triggers a "Manual Review" dialog in the UI.
    6. User rejects the plan; agent is restricted from executing any steps.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant A as AI Agent
        participant P as PVN Middleware
        participant S as Semantic Analyzer
        participant H as Hardware Enclave (TPM)
        participant U as User (UI)

        A->>P: Submit Plan (JSON/MD)
        P->>S: Analyze Sequence
        S-->>P: Risk Score & Violations
        alt Risk == Low
            P->>H: Sign Plan Approval
            H-->>P: Approval Token
            P-->>A: Authorized (Token)
        else Risk == High
            P->>U: Request Manual Review
            U-->>P: Reject/Approve
            P-->>A: Restriction/Authorization
        end
    ```
* **APIs / Interfaces:**
    * `POST /v1/plan/verify`: Accepts an array of tool-call proposals.
    * `GET /v1/plan/status/{plan_id}`: Polls for manual review status.
* **Data Storage/State:**
    * Plans are stored in the **Universal Episodic Graph** with a status of `PENDING`, `AUTHORIZED`, or `REJECTED`.

## 5. Alternatives Considered
* **Individual Tool Gating**: Rejected because it misses the "big picture" logic of multi-step attacks.
* **LLM-Based Verification**: Rejected as primary method due to "Mirroring" risks (a second LLM might be fooled by the same prompt injection). Hardware-attested semantic rules are the primary path.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Approval tokens are hardware-bound and expire after the plan is executed or after a short TTL.
* **Observability**: All plan verification attempts, risk scores, and user decisions are logged to the `mcpany` audit log.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
