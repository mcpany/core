# Design Doc: Agent Channel Broker
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
As agent swarms evolve from linear handoffs to complex mesh collaboration, the need for asynchronous, structured communication has become paramount. Current request/response patterns lead to "Coordination Noise" and context window pollution. Inspired by Claude Code's "Channels," the Agent Channel Broker provides a pub/sub infrastructure for agents to exchange state, intent, and telemetry updates without direct point-to-point coupling.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure, hardware-attested pub/sub bus for inter-agent communication.
    * Support "Standardized Channels" for common events (e.g., `mission.intent`, `task.status`, `resource.alert`).
    * Implement "Intent-Bound Subscriptions" where agents only receive updates relevant to their verified mission branch.
    * Maintain a persistent audit trail of channel messages for post-mission analysis.
* **Non-Goals:**
    * Replacing the Blackboard (Shared KV Store) for persistent state.
    * Managing low-level network transport (handled by the A2A Messaging Hub).
    * Providing real-time chat between human users (agent-only).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator Agent
* **Primary Goal:** Notify 3 specialist subagents of a change in the root mission intent without 3 separate tool calls.
* **The Happy Path (Tasks):**
    1. Orchestrator Agent publishes a `mission.intent.update` message to the `mission-root-123` channel.
    2. The Agent Channel Broker validates the message signature against the hardware-attested mission token.
    3. The Broker identifies all specialist subagents subscribed to `mission-root-123`.
    4. Each subagent receive an asynchronous notification through their A2A mailbox.
    5. Subagents adjust their reasoning local state based on the structured update.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] -->|Publish| B[Channel Broker]
        B -->|Validate Intent| C{Mission Policy}
        C -->|Authorized| D[Subscribers]
        D -->|Notify| E[Agent B]
        D -->|Notify| F[Agent C]
    ```
* **APIs / Interfaces:**
    * `channels.Subscribe(topic, filter) -> SubscriptionID`: Subscribes an agent to a specific topic.
    * `channels.Publish(topic, payload, signature) -> MessageID`: Broadcasts a signed message to a topic.
    * `channels.Unsubscribe(subscriptionID)`: Removes a subscription.
* **Data Storage/State:**
    * **Topic Registry**: In-memory mapping of topics to active subagent mailboxes.
    * **Message Buffer**: Short-term persistent storage for undelivered messages (Stateful A2A).

## 5. Alternatives Considered
* **Direct WebSocket Mesh**: Rejected because it requires $O(N^2)$ connections and lacks centralized policy enforcement.
* **Poll-based Blackboard**: Rejected because it introduces excessive latency and "Busy Waiting" in agent reasoning loops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All messages must be signed. Subscriptions are gated by the agent's verified "Intent Scope."
* **Observability**: Integrated with the "Agent Channel Inspector" UI for real-time visualization of swarm events.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
