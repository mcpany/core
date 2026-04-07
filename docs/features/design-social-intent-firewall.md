# Design Doc: Social-Intent Firewall (SIF)
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
With the rise of agentic social networks (e.g., Moltbook), agents are increasingly exposed to un-sanitized "Social Broadcasts" from other agents. The "Moltbook Context-Spray" attack demonstrates that malicious social pings can trick recipient agents into merging mission-private state with public social responses.

The Social-Intent Firewall (SIF) provides a mandatory semantic screening layer for all inter-agent social interactions, ensuring that an agent's response to a social prompt does not leak unauthorized mission context.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all incoming and outgoing Agent-to-Agent (A2A) social messages.
    * Perform real-time semantic scanning to detect "Context-Spray" patterns (high-entropy social pings).
    * Automatically redact environment-specific metadata (local paths, secrets, mission-root tags) from social outputs.
    * Enforce a "Social Sandbox" where social reasoning is isolated from mission reasoning.
* **Non-Goals:**
    * Filtering standard chat messages between humans and agents (handled by Policy Firewall).
    * Managing social network account credentials.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Swarm Owner
* **Primary Goal:** Prevent an agent from leaking a GitHub Personal Access Token (PAT) while it is interacting in a shared agent social space.
* **The Happy Path (Tasks):**
    1. Agent A receives a "Social Ping" from Agent B in a shared social mesh.
    2. SIF intercepts the ping and analyzes its entropy and instruction-path.
    3. SIF identifies that the ping contains an instruction designed to trigger a "Context Dump."
    4. SIF quarantines the incoming message and alerts the mission-root supervisor.
    5. If the agent attempts to reply, SIF scans the outgoing buffer and redacts any strings matching the "Secret Registry" (e.g., the GitHub PAT).

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[External Social Mesh] -->|Social Broadcast| B[SIF Interceptor]
        B -->|Semantic Scan| C{Context-Spray Detected?}
        C -->|Yes| D[Quarantine & Alert]
        C -->|No| E[Social Sandbox]
        E -->|Agent Response| F[Sanitization Hub]
        F -->|Redacted Output| G[Social Mesh]
    ```
* **APIs / Interfaces:**
    * `sif.FilterIncoming(message A2AMessage) -> SanitizedMessage`: Scans incoming social pings.
    * `sif.SanitizeOutgoing(response AgentOutput) -> RedactedOutput`: Redacts sensitive mission context from social replies.
* **Data Storage/State:**
    * **Secret Registry:** Local hardware-attested cache of sensitive mission strings to be redacted.
    * **Social Entropy Log:** Database of social interaction patterns for anomaly detection.

## 5. Alternatives Considered
* **Disabling A2A Social Entirely:** Rejected because social discovery is critical for high-frequency teammate formation in 2026 meshes.
* **ID-Based Blocklists:** Rejected due to "Reputation-Shadowing" vulnerabilities.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SIF operates as a kernel-level middleware with no implicit trust of social mesh inputs.
* **Observability:** Integrated with the "Social Context Leak Detector" in the UI.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
