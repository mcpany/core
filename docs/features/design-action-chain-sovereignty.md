# Design Doc: Action-Chain Sovereignty (ACS) Monitor
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
As autonomous agents execute longer and more complex sequences of actions (Action-Chains), the risk of "Machine-Speed Cascading Failures" increases. A single compromised or malfunctioning agent can initiate a chain of system-wide changes faster than any human operator can intervene. MCP Any needs a mechanism to validate the entire sequence of automated workflows against the cryptographically signed "Mission-Root" intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time monitoring of automated agent action sequences.
    * Validate each step in the action-chain against the mission-root manifest.
    * Provide automated "Chain Interdiction" if a sequence diverges from authorized patterns.
* **Non-Goals:**
    * Replacing existing per-tool authorization (this adds a temporal/sequential layer).
    * Managing the business logic of the workflow itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Engineer
* **Primary Goal:** Prevent a GitHub triage bot from escalating into a CI/CD build cache poisoning attack after ingesting a malicious issue title.
* **The Happy Path (Tasks):**
    1. Agent ingests a new GitHub issue and starts a "Triage Action-Chain."
    2. The chain includes `read_issue`, `label_issue`, and `comment_issue`. These are authorized.
    3. The agent (influenced by malicious metadata) attempts to append `edit_build_config` to the chain.
    4. ACS Monitor detects that `edit_build_config` is not a valid successor in the "Triage" mission scope.
    5. ACS Monitor interdicts the call and revokes the agent's session tokens.
    6. Engineer is notified of the "Chain Divergence" attempt.

## 4. Design & Architecture
* **System Flow:**
    * Agents register an "Action Manifest" at mission start.
    * Every action request is tagged with a `Chain-ID` and `Sequence-Index`.
    * The ACS Middleware evaluates the request against the `Mission-Graph` (state machine of authorized actions).
* **APIs / Interfaces:**
    * `acs.v1.RegisterChain(mission_id, graph_definition)`: Defines the authorized state machine for a mission.
    * `acs.v1.ValidateAction(chain_id, action_metadata)`: Middleware hook for per-action validation.
* **Data Storage/State:**
    * Action histories are stored in the **Blackboard**, indexed by `Chain-ID`.

## 5. Alternatives Considered
* **Stateless Policy Checks:** Rejected as they cannot detect "Action Smuggling" where a series of individually safe actions leads to an unsafe state.
* **Human-in-the-Loop for Every Step:** Rejected due to the performance requirement of machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Chain manifests must be hardware-attested by the user at mission start.
* **Observability:** Provides a "Chain Visualizer" in the UI to show active vs. predicted action paths.

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
