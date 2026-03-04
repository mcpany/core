# Design Doc: Contextual Tool-Level MFA (Step-Up Auth)
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
Existing MCP adapters often require all tool calls to be approved (HITL) OR all to be automated. This is insufficient for sensitive production environments where some tools (like `filesystem:delete`) are too risky for autonomous execution, even if the agent is "trusted."

MCP Any needs a "Step-Up" authentication protocol that can trigger a real-time user approval flow on a mobile device or hardware token for specific, sensitive tool calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Trigger MFA (e.g., Push Notification, WebAuthn) for high-risk tool calls.
    * Provide a standardized way to define "MFA-Required" tools in `config.yaml`.
    * Suspend the tool call execution until the user approves.
    * Support "Time-Bound" approvals (e.g., Approve for 30 minutes).
* **Non-Goals:**
    * Replace agent login MFA.
    * Handle identity management (rely on OIDC or existing auth).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cloud Infrastructure Manager
* **Primary Goal:** Ensure an agent cannot terminate an EC2 instance without a direct confirmation on the manager's phone.
* **The Happy Path (Tasks):**
    1. The agent identifies a need to terminate a stale instance and calls `aws:terminate_instance`.
    2. MCP Any identifies `aws:terminate_instance` as an "MFA-Required" tool.
    3. MCP Any suspends the tool call and sends a push notification to the manager's registered device.
    4. The manager reviews the request (with full tool arguments) on their phone and hits "Approve."
    5. MCP Any resumes the tool call and completes the execution.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Tool Call (aws:terminate_instance)` -> `MFA Hook (Middleware)` -> `Is MFA Required? (Yes)`
    `MFA Hook` -> `Send Push Notification (Mobile App/Web)` -> `Suspend Call (Context.Pause)`
    `User` -> `Approves on Phone` -> `Resume Call (Context.Resume)` -> `Upstream Call`
* **APIs / Interfaces:**
    * `/api/mfa/approve`: To be called by the user's mobile app or a secure web portal.
    * `/api/mfa/deny`: To cancel the request.
* **Data Storage/State:**
    * Suspended calls are stored in a persistent "Waiting Room" (SQLite).

## 5. Alternatives Considered
* **CLI Approval:** Prompting the user in the CLI. Rejected because it requires the user to be active on their terminal, which isn't always the case for autonomous agents.
* **Policy-Based Deny:** Just blocking the tool. Rejected because it prevents legitimate tasks from being completed autonomously with oversight.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Approval tokens must be cryptographically signed by the mobile device.
* **Observability:** All MFA approvals and denials are logged for SOC2 audit compliance.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
