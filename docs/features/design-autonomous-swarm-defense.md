# Design Doc: Autonomous Swarm Defense (ASD)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the rise of "AI Swarm Attacks" (Hivenet) and the increasing complexity of multi-agent systems, traditional Human-in-the-Loop (HITL) security is becoming a bottleneck and a point of failure. Swarms can execute thousands of tool calls in seconds, far outstripping human reaction time. Autonomous Swarm Defense (ASD) is a low-latency, high-throughput security middleware for MCP Any that identifies and terminates malicious swarm patterns in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide millisecond-latency evaluation of tool call sequences.
    * Identify anomalous swarm behaviors (e.g., rapid lateral tool movement, unauthorized recursive spawning).
    * Automatically terminate suspicious tool-chains without waiting for human approval.
    * Implement "Intent-Bound Traceability" to link every sub-agent action back to a parent intent.
* **Non-Goals:**
    * Replacing general-purpose LLM safety filters (focus is on tool-use and swarm behavior).
    * Being a general-purpose antivirus (focus is on agentic actions).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Operator
* **Primary Goal:** Protect internal infrastructure from a compromised or rogue AI swarm that attempts to rapidly exfiltrate data.
* **The Happy Path (Tasks):**
    1. A parent agent spawns 50 sub-agents to "analyze the network."
    2. Sub-agents begin calling `list_directory` and `read_file` across multiple sensitive volumes in parallel.
    3. ASD detects a "High-Velocity Lateral Movement" pattern that exceeds the established "Intent Scope."
    4. ASD immediately revokes the session tokens for all sub-agents and freezes the parent process.
    5. ASD sends an alert to the user with a cryptographic trace of the actions taken and the policy violated.

## 4. Design & Architecture
* **System Flow:**
    `Agent(s)` -> `MCP Gateway` -> `ASD Middleware` -> `Policy Engine (CEL/Rego)` -> `Tools`
    1. **Observation**: ASD intercepts every tool request and records it in a high-speed in-memory "Swarm Graph."
    2. **Pattern Matching**: The engine runs Common Expression Language (CEL) rules against the graph (e.g., `graph.velocity > 100 && graph.entropy > threshold`).
    3. **Enforcement**: If a rule triggers, ASD issues a "Kill" signal to the Gateway, which returns an error to the agents and invalidates their session.
* **APIs / Interfaces:**
    * `SwarmGraph.Record(event)`: Internal interface for tracking tool calls.
    * `ASD.UpdatePolicy(ruleset)`: API for updating the autonomous rules.
* **Data Storage/State:**
    * **Redis (In-Memory)**: Stores the real-time swarm graph for sub-millisecond lookups.
    * **Audit Log (Persistent)**: Records every "Kill" event for forensics.

## 5. Alternatives Considered
* **Manual HITL**: Rejected as too slow for swarm-speed attacks.
* **Static Permission Scoping**: Too rigid for dynamic swarms; doesn't account for behavioral velocity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: ASD itself must be isolated and protected from agent manipulation.
* **Observability**: Real-time dashboard for swarm health and defense status.
* **Latency**: ASD must add < 5ms to the tool call round-trip.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
