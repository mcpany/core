# Feature Inventory & Backlog

## Core Feature List

### 1. Universal Adapters
- [x] HTTP/REST Adapter
- [x] gRPC Adapter (Dynamic Discovery)
- [x] Command Adapter (Basic Execution)
- [x] Filesystem Adapter

### 2. Security & Policy
- [x] API Key Authentication
- [x] Environment Variable Redaction
- [ ] **[Priority] Zero Trust Policy Firewall (Rego/CEL):** Granular control over tool calls based on context, user, and payload.
- [ ] **[Priority] Isolated Execution Enclaves:** Docker-bound execution for command adapters.

### 3. Agentic Orchestration (Universal Agent Bus)
- [ ] **[Priority] Recursive Context Protocol:** Standardized headers for context inheritance between agents and subagents.
- [ ] **[Priority] Shared Key-Value Store (Blackboard):** SQLite-backed shared state for agent swarms.
- [ ] **[Priority] Isolated Inter-Agent Communication:** Using Docker-bound named pipes for secure subagent routing.

### 4. Observability & UX
- [x] Real-time Health Dashboard
- [x] Tool Usage History
- [ ] **[Priority] HITL Middleware:** Human-in-the-loop approval flow for sensitive operations.
- [ ] Interactive Trace Visualizer (Sequence Diagrams)

## Backlog Grooming: [2026-02-19]

### Proposed New Features
1. **Isolated Inter-Agent Communication:**
   - **Context:** Today's market sync revealed exploit patterns in agent routing.
   - **Description:** Replace local HTTP tunneling with isolated Docker-bound named pipes for inter-agent comms to prevent unauthorized host-level access.
   - **Priority:** High (P0)

2. **Zero Trust Policy Firewall (Rego/CEL):**
   - **Context:** Critical vulnerabilities in OpenClaw (CVE-2026-25253) show that agents need strictly enforced boundaries.
   - **Description:** A declarative policy engine that evaluates every tool call against a Rego policy before execution.
   - **Priority:** High (P0)

3. **Recursive Context Protocol:**
   - **Description:** Solve the "subagent configuration pain" by standardizing how context is passed down the agent chain.
   - **Priority:** Medium (P1)
