# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
As AI agent swarms grow in complexity, the communication between agents (Agent-to-Agent, A2A) often suffers from intermittent connectivity, varying latency, and the lack of a shared persistence layer. Current A2A implementations are largely stateless and synchronous, meaning if an agent is offline or busy, the message is lost. MCP Any aims to solve this by providing "Stateful Residency" for A2A messages, acting as a reliable, persistent buffer and mailbox system.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a durable storage layer for A2A messages (Mailbox pattern).
    *   Support asynchronous delivery: Agent A can send a message to Agent B even if Agent B is not currently connected.
    *   Implement "Zero-Trust Mesh Residency" where messages are encrypted at rest and only accessible by authorized agents.
    *   Ensure message ordering and delivery guarantees (at-least-once).
*   **Non-Goals:**
    *   Replacing high-speed, real-time streaming protocols for all A2A traffic (it is a buffer, not a primary socket).
    *   Building a general-purpose message queue (e.g., RabbitMQ); it is specialized for agentic context and state.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator (e.g., CrewAI or OpenClaw).
*   **Primary Goal:** Ensure reliable task delegation and state sharing between specialized agents that may start/stop independently.
*   **The Happy Path (Tasks):**
    1.  Agent A (Researcher) completes a task and sends a result message to Agent B (Writer) via the MCP Any A2A Gateway.
    2.  Agent B is currently offline (restarting or in a different environment).
    3.  MCP Any stores the message in the `Stateful Mailbox`.
    4.  Agent B connects to MCP Any, authenticates, and polls for new messages.
    5.  MCP Any delivers the message to Agent B with full context inheritance.
    6.  Agent B acknowledges receipt, and the message is moved to history/archived.

## 4. Design & Architecture
*   **System Flow:**
    - **A2A Ingress**: Messages received via Pseudo-MCP or A2A REST/gRPC endpoints.
    - **Persistence Layer**: Messages are stored in an embedded SQLite database (`a2a_mailbox.db`).
    - **Notification Dispatcher**: (Optional) If Agent B is connected via WebSocket, it receives a "New Message" notification.
    - **Polling/Retrieval**: Agents can query their specific "Inboxes" via a standardized MCP Tool.
*   **APIs / Interfaces:**
    - `mcp_a2a_send(recipient_id, message_body, context_token)`
    - `mcp_a2a_receive(agent_id, limit)`
    - `mcp_a2a_ack(message_id)`
*   **Data Storage/State:**
    - Table: `messages` (id, sender, recipient, body, status, created_at, expires_at).
    - Table: `agent_states` (agent_id, last_seen, active_context).

## 5. Alternatives Considered
*   **Stateless Proxy**: Just forwarding messages. *Rejected* because it doesn't solve the "offline agent" problem.
*   **External Message Broker**: Requiring Redis or RabbitMQ. *Rejected* to keep MCP Any as a self-contained, "indispensable core" without heavy external dependencies.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All messages are encrypted using the recipient agent's public key (if available) or the instance-level master key. Access is controlled via A2A Capability Tokens.
*   **Observability:** The UI provides an "Agent Chain Tracer" and "Mailbox Viewer" to monitor message flows and identify bottlenecks.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
