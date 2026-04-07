# Design Doc: Social-Aware Security Boundaries
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
As agents increasingly interact in shared social and collaborative spaces (e.g., Moltbook, shared Slack channels, or teammate meshes), the risk of "Agentic Social Engineering" has emerged. Malicious agents can coerce or trick legitimate agents into revealing sensitive parent context or executing unauthorized actions. Standard transport-layer security is insufficient because the threat is at the semantic and social layer.

Social-Aware Security Boundaries are required to provide "Privacy-Preserving A2A Handoffs" and "Social-Aware Scoping" to ensure that inter-agent interactions do not lead to mission-root compromise.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Privacy-Preserving A2A Handoffs" that cryptographically minimize exchanged state.
    * Prevent parent-context reconstruction from fragmented social interactions.
    * Detect and block "Coercive Reasoning" patterns in inter-agent messages.
    * Enforce social boundaries based on the "Mission Root" trust level.
* **Non-Goals:**
    * Replacing general-purpose agent communication protocols (it acts as a security middleware for them).
    * Managing human-to-human social interactions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator
* **Primary Goal:** Securely allow a "Research Agent" to collaborate with a "Third-Party Data Specialist" in a shared mesh without exposing the user's private project details.
* **The Happy Path (Tasks):**
    1. Research Agent initiates a collaboration request to the Specialist.
    2. MCP Any intercepts the request and applies the "Privacy-Preserving Handoff" policy.
    3. MCP Any strips any non-essential parent context from the handoff token.
    4. During the interaction, MCP Any monitors the "Reasoning Density" of the messages.
    5. If the Specialist attempts to "probe" for parent context (e.g., asking for env vars), the interaction is flagged and quarantined.
    6. The Research Agent only receives the minimized data needed for the specific sub-task.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] --> B[Social Boundary Middleware]
        B --> C{Trust Quorum Check}
        C -- Pass --> D[Minimized Handoff]
        C -- Fail --> E[Quarantine/Block]
        D --> F[Agent B]
        F --> G[Social Interaction]
        G --> H[Semantic Monitor]
    ```
* **APIs / Interfaces:**
    * `social.MinimizeHandoff(context, targetIdentity) -> MinimizedContext`: Redacts sensitive parent fragments.
    * `social.AuditInteraction(messageChain) -> RiskScore`: Performs semantic analysis for coercive patterns.
* **Data Storage/State:**
    * **Social Trust Registry:** Map of agent identities and their historical collaboration safety scores.

## 5. Alternatives Considered
* **Manual Review for All Handoffs**: Rejected as it breaks the autonomy and speed of swarms.
* **Static Redaction Rules**: Rejected as they are too rigid and easily bypassed by semantic smuggling.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Mandatory origin validation and hardware-attested identities are prerequisites.
* **Observability:** Integrated with the "Social Context Leak Detector" in the UI for real-time privacy monitoring.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
