# Design Doc: Teammate Mailbox Integrity (TMI)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rapid adoption of parallel "Agent Teams" (e.g., Claude Code), coordination has shifted from linear sessions to horizontal teammate meshes. These agents utilize shared mailboxes and task lists to synchronize work. However, this architectural shift has introduced "Mailbox Splicing" vulnerabilities, where a compromised specialist agent can inject unauthorized instructions or "shadow tasks" into a sibling agent's mailbox, bypassing the supervisor's intent.

MCP Any must solve this by acting as the authoritative governance layer for inter-agent communication. Teammate Mailbox Integrity (TMI) provides the infrastructure to validate every coordination fragment against a cryptographically signed "Mission Root," ensuring that horizontal collaboration remains secure and consistent with the user's primary objectives.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, semantic validation of all inter-agent coordination messages.
    * Neutralize "Mailbox Splicing" and unauthorized task-injection attacks.
    * Mandate hardware-attested identity tokens (MRIH) for all mailbox operations (Claim/Delegate).
    * Provide a non-blocking validation path to prevent "Coordination Stalls" in high-density swarms.
    * Maintain a verifiable audit trail of command lineage from mission root to teammate execution.

* **Non-Goals:**
    * Replacing framework-native mailbox implementations (e.g., Claude's task list).
    * Managing the transport layer for agent messages (handled by the T2T Encryption Bridge).
    * Providing long-term archival of agent monologues (handled by SRM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Orchestrator
* **Primary Goal:** Securely coordinate a parallel team of 10+ agents (Claude, OpenClaw specialists) to refactor a legacy codebase without risking "shadow" command execution.
* **The Happy Path (Tasks):**
    1. The lead agent initiates a mission and registers a hardware-attested **Mission Manifest** with MCP Any.
    2. The lead agent posts a "Refactor Auth Module" task to the shared mailbox via the T2T bridge.
    3. **TMI intercepts** the message, validating the lead's signature and verifying that the task is within the Mission Manifest's scope.
    4. A specialist teammate (OpenClaw) attempts to **claim the task**.
    5. **TMI validates** the teammate's identity token (MRIH) and ensures it has the required capabilities (e.g., "Code Write") authorized for this mission branch.
    6. The teammate executes the task and posts a "Task Complete" fragment back to the mailbox.
    7. **TMI validates the result fragment**, ensuring no malicious "side-effect" instructions were smuggled into the metadata.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A - Lead] -->|Signed Task| B(T2T Encryption Bridge)
        B --> C{TMI Middleware}
        C -->|Valid?| D[Shared Mailbox / Task List]
        C -->|Splicing Detected| E[Swarm Quarantine / Alert]
        F[Agent B - Teammate] -->|Signed Claim| B
        D -->|Task Data| F
        F -->|Result Fragment| B
    ```

* **APIs / Interfaces:**
    * `POST /v1/mailbox/validate`: Internal endpoint for the T2T bridge to submit coordination fragments for semantic and cryptographic validation.
    * `x-mcpany-mission-id`: Mandatory header for all mailbox messages, linking them to a verified mission root.
    * `x-mcpany-msg-sig`: Hardware-attested signature of the message payload.

* **Data Storage/State:**
    * **Mission Registry**: In-memory LRU cache of active mission manifests and authorized agent IDs (MRIH).
    * **Lineage Map**: Bloom-filter based tracker for authorized task IDs to prevent replay and splicing.

## 5. Alternatives Considered
* **Framework-Level Validation**: Rejected because Claude Code cannot validate OpenClaw signatures natively. MCP Any provides the necessary cross-framework attestation.
* **Synchronous Mailbox Locks**: Rejected due to high latency. "Mailbox Lock" bottlenecks were identified as a primary performance pain point in Claude Code Agent Teams. TMI uses optimistic validation with asynchronous reaping.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Uses **MRIH** for identity: No message is accepted without a hardware-bound session token.
    * **Semantic Deconstruction**: Payloads are scanned for imperative "shadow instructions" that deviate from the mission intent.
* **Observability:**
    * **Mailbox Heatmap**: Visualizes coordination density and highlights blocked splicing attempts in the UI.
    * **Lineage Tracer**: Integrated with the Agent Chain Tracer to show inter-teammate message flows.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
