# Design Doc: A2UI Native Gateway
**Status:** Draft
**Created:** 2026-04-21

## 1. Context and Scope
As AI agent swarms become more autonomous and specialized, they often need to present complex information or request user input in ways that simple text cannot handle. The Agent-to-User Interface (A2UI) protocol has emerged to solve this by allowing agents to generate secure, interactive UI fragments.

MCP Any, as the universal agent infrastructure, is uniquely positioned to act as the A2UI Gateway. It will provide the bridge between remote agents (communicating via A2A or MCP) and the local user's workspace, ensuring that UI components are rendered securely, isolated from the host environment, and verified against the agent's identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a secure proxy for A2UI messages between agents and the UI frontend.
    * Provide a sandboxed rendering environment for agent-generated UI fragments.
    * Enforce origin-based validation for all UI-related messages to prevent cross-site scripting (XSS) and injection.
    * Support bi-directional state synchronization between the agent and its surfaced UI.
* **Non-Goals:**
    * Defining the visual styling or component library of the UI (A2UI is protocol-agnostic regarding specific frameworks).
    * Executing arbitrary JavaScript from agents on the host (all UI logic must be handled via declarative state or pre-defined secure components).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Developer
* **Primary Goal:** Surface a secure interactive dashboard from a remote specialist agent to a local developer workspace.
* **The Happy Path (Tasks):**
    1. A specialist agent (e.g., a Database Tuner) identifies a need for user feedback on a complex configuration.
    2. The agent sends an A2UI "Manifest" message containing the UI state and interactive elements via the MCP Any A2A Hub.
    3. The A2UI Gateway validates the agent's identity and signs the UI manifest.
    4. The MCP Any UI frontend receives the manifest and renders the sandboxed component.
    5. The user interacts with the UI (e.g., toggling a setting).
    6. The A2UI Gateway captures the interaction event and routes it back to the agent as a state update.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> (A2A/MCP) -> [A2UI Gateway (Server)] -> (WebSocket) -> [Sandboxed UI Component (Frontend)]`
* **APIs / Interfaces:**
    * `/v1/a2ui/register`: Allows agents to declare their UI capabilities and schemas.
    * `/v1/a2ui/stream`: A WebSocket endpoint for real-time UI state synchronization.
    * `A2UI Message Object`: Standardized JSON schema for UI state, events, and lifecycle signals.
* **Data Storage/State:**
    * UI state is maintained in-memory within the Gateway session, with periodic checkpoints in the Blackboard.

## 5. Alternatives Considered
* **Direct Agent-to-Browser Comms:** Rejected due to the OpenClaw local trust crisis (CVE-2026-25253). Bypassing the gateway would expose the browser to unvalidated local listeners.
* **Iframe-based Rendering:** Considered but rejected in favor of a more granular "Secure Component" model to prevent parent-page exfiltration and ensure a native feel.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All A2UI manifests must be cryptographically signed by the agent. The Gateway performs strict schema validation to prevent "UI Injection" attacks.
* **Observability:** Every UI interaction is logged in the "A2UI Audit Trail," allowing users to see exactly what state was proposed and what they approved.

## 7. Evolutionary Changelog
* **2026-04-21:** Initial Document Creation.
* **2026-04-22:** Update: Resolving WebSocket State Desync.
    * **Context:** Market research revealed state desync when switching between Normal and Adaptive reasoning modes in OpenClaw v2026.3.1.
    * **Architecture Adjustment:**
        * Implementing "State Versioning" in the A2UI Stream. Every UI manifest now carries a `reasoning_epoch` ID.
        * The Gateway will buffer and "Fast-Forward" UI updates if an epoch switch is detected, ensuring the visual state matches the current reasoning effort.
    * **Security Impact:** Prevents "UI Shadowing" where an agent in a lower-effort mode could be manipulated by stale UI state from a high-effort reasoning branch.
