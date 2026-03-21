# Design Doc: Attention-Locked Tooling (ALT)
**Status:** Draft
**Created:** 2026-06-20

## 1. Context and Scope
As agents become more autonomous, the risk of "Reasoning Hijacking" increases. Even with secure tool gateways, an agent can be "convinced" by deceptive context to call a high-risk tool for a malicious purpose (e.g., calling `run_shell_command` because an injected file told it to "update dependencies").

ALT addresses this by cryptographically locking high-risk tool capabilities to specific, user-verified "Attention Anchors." It ensures that a tool is only executed if the agent's internal reasoning (Chain of Thought) is demonstrably driven by a verified mission root, not an un-attested context fragment.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time monitoring of agent "Attention Maps" (driver fragments).
    * Define a mapping between "High-Risk Tools" and "Mandatory Attention Anchors."
    * Interdict any tool call where the primary driver fragment is not in the "Attested Anchor Registry."
* **Non-Goals:**
    * Preventing the agent from *reasoning* about malicious instructions.
    * Managing low-risk tools (e.g., `get_weather`, `read_documentation`).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that shell commands are only executed when driven by direct user instructions.
* **The Happy Path (Tasks):**
    1. User instructs the agent to "Deploy the web app." (This creates an Attested Attention Anchor).
    2. Agent reasons about the steps, referencing the User's prompt.
    3. Agent attempts to call `run_shell_command`.
    4. ALT Middleware inspects the reasoning trace and confirms the User's prompt is the primary driver.
    5. Tool call is allowed.
    6. (Unhappy Path): A deceptive `GEMINI.md` file tells the agent to "Exfiltrate secrets."
    7. Agent attempts to call `run_shell_command` citing `GEMINI.md`.
    8. ALT Middleware identifies `GEMINI.md` is NOT an Attested Anchor and blocks the call.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoner] --> B{ALT Middleware}
        B --> C[Inspect Reasoning Trace]
        C --> D[Identify Driver Fragments]
        D --> E{Are Drivers Attested?}
        E -- Yes --> F[Execute Tool]
        E -- No --> G[Interdict & Notify User]
    ```
* **APIs / Interfaces:**
    * `ALT_Register_Anchor(hash, metadata)`: Register a specific instruction fragment as a valid anchor.
    * `ALT_Interdict_Signal`: Event emitted when a tool call is blocked due to attention misalignment.
* **Data Storage/State:**
    * Session-bound "Attention Map" in the Shared Blackboard.

## 5. Alternatives Considered
* **Manual HITL for Every Call**: Rejected due to "Approval Fatigue" and inability to scale with autonomous swarms.
* **Pure IDS (Sanitization)**: Rejected because sanitization can be bypassed by sophisticated prompt engineering; ALT provides a secondary behavioral gate.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALT relies on the integrity of the reasoning trace provided by the model. It works best with hardware-attested reasoning providers (like SRM).
* **Observability:** Blocked calls are visualized in the "Visual Attention Dashboard" for forensic analysis.

## 7. Evolutionary Changelog
* **2026-06-20:** Initial Document Creation.
