# Design Doc: Environment Bridging Middleware

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
With the rise of cloud-sandboxed agents (e.g., Claude Code, Gemini CLI's restricted environments), there is a widening gap between these restricted agents and local development tools. MCP Any needs to bridge this "Local-to-Cloud" gap by providing a secure, low-latency middleware that allows cloud agents to execute tools in the user's local environment without compromising the security of the host.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure, authenticated tunnel for cloud-sandboxed agents to call local MCP tools.
    * Synchronize minimal state (project directory, environment variables) required for tool execution.
    * Enforce a "Cloud-to-Local" security policy using Zero-Trust principles.
    * Minimize latency overhead for remote tool calls.
* **Non-Goals:**
    * Providing general-purpose remote access to the host machine.
    * Managing the lifecycle of the cloud agent itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Cloud-native AI Developer.
* **Primary Goal:** Use a cloud-sandboxed agent (e.g., Anthropic's Claude Code) to run local `git` or `npm` commands via MCP Any.
* **The Happy Path (Tasks):**
    1. User starts MCP Any locally with the `env-bridge` enabled.
    2. User initiates a cloud-sandboxed agent and provides the MCP Any bridge URL and an authentication token.
    3. The cloud agent calls a local tool (e.g., `git status`) via the bridge.
    4. MCP Any validates the token and the tool call against the local security policy.
    5. MCP Any executes the tool locally and returns the result to the cloud agent.

## 4. Design & Architecture
* **System Flow:**
    - **Authentication**: Uses Ed25519-signed JWTs for mutual authentication between the cloud agent and the local MCP Any node.
    - **Tunneling**: Implements a secure WebSocket or gRPC tunnel (likely over a relay or authenticated endpoint).
    - **State Sync**: Uses the `Recursive Context Protocol` to pass session-scoped state from the cloud agent to the local tool.
* **APIs / Interfaces:**
    - `POST /bridge/authorize`: Generates a one-time authorization token for a cloud agent.
    - `WS /bridge/tunnel`: The secure websocket endpoint for bidirectional tool calling.
* **Data Storage/State:** Bridge session metadata is stored in memory; persistent state is handled by the `Shared KV Store`.

## 5. Alternatives Considered
* **Local Proxying via Ngrok**: Rejected due to lack of granular MCP-aware security policies and high friction.
* **Direct SSH Tunneling**: Rejected as it is too complex for most users and lacks the necessary protocol-level integration.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All "Cloud-to-Local" calls are subject to the `Policy Firewall` and must be explicitly allowed by the user or a predefined "Safe-Tool" list.
* **Observability:** The UI will display a "Bridge Status" indicator, showing active cloud connections and the latency of remote tool calls.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
