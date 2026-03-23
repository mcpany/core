# Design Doc: Secure Coordination Bus
**Status:** Draft | In Review | Approved
**Created:** 2026-05-16

## 1. Context and Scope
With the launch of Claude Code's "Agent Teams," the industry is moving toward parallel, multi-agent swarms. These teams must coordinate and share state (e.g., via the Blackboard). We have identified "Mailbox Injection" as a catastrophic new attack vector where a compromised specialist agent injects malicious coordination messages into a sibling's inbox to hijack its reasoning or exfiltrate data. MCP Any needs a secure coordination layer that cryptographically signs every message in the swarm's bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographically signed transport for teammate-to-teammate messages.
    * Enable high-speed "Snapshot-and-Merge" state reconciliation.
    * Neutralize "Identity Shadowing" attacks within the swarm.
* **Non-Goals:**
    * Replace the agent's internal message passing logic.
    * Provide a general-purpose chat protocol for human users.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 parallel agents without exposing local env vars or risking mailbox injection.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes 3 agents (A1, A2, A3) on the Secure Coordination Bus.
    2. Agent A1 fetches a git hash and broadcasts it to the bus.
    3. The message is signed with A1's session-bound identity.
    4. Agent A2 receives the message and verifies the signature using MCP Any's internal trust broker.
    5. A malicious subagent attempts to inject a spoofed message to A3.
    6. Agent A3's inbox validator blocks the message because it lacks a valid signature from an authorized teammate.

## 4. Design & Architecture
* **System Flow:**
    1. **Teammate Registration**: Agents exchange public keys during initialization.
    2. **Message Signing**: Every broadcast message is signed with a private key bound to the agent's mission-token.
    3. **Inbox Validation**: MCP Any acts as the "Inbox Broker," verifying signatures before delivering messages to agents.
    4. **Snapshot-and-Merge**: Parallel state branches are reconciled using signed "State Diff" objects.
* **APIs / Interfaces:**
    * `POST /v1/swarm/broadcast`: Sending a signed message to the bus.
    * `GET /v1/swarm/inbox`: Retrieving verified teammate messages.
* **Data Storage/State:** Teammate public keys and recent message hashes are stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Implicit Trust on Localhost**: Rejected because browser-based attackers can bridge into the agent's control plane via unauthenticated loopback listeners.
* **Centralized Orchestrator Only**: Rejected because it creates a performance bottleneck and doesn't scale to autonomous peer-to-peer swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** We enforce "Auth-at-the-Pipe" for all inter-agent communication, ensuring that even if one agent is compromised, the bus remains secure.
* **Observability:** Swarm-wide message flows are logged in a cryptographically signed audit trail.

## 7. Evolutionary Changelog
* **2026-05-16:** Initial Document Creation.
