# Feature Inventory: MCP Any

## Current Backlog (P0/P1)
- **Policy Firewall**: Rego/CEL based hooking for tool calls.
- **HITL Middleware**: Suspension protocol for user approval flows.
- **Recursive Context Protocol**: Standardized headers for subagent inheritance.
- **Shared KV Store**: Embedded SQLite "Blackboard" tool for agents.

---

## Evolution: [2026-02-23] Updates

### Proposed Additions
- **Environment Bridging Middleware**: (P1) Bridge between cloud-sandboxed agents (e.g., Claude Code Sandbox) and local MCP Any tools. Enables seamless state transfer.
- **Machine-Checkable Security Contracts**: (P1) Declarative security models for tools that can be verified by automated agents (inspired by OpenClaw).
- **Zero-Trust Subagent Scoping**: (P0) Capability-based tokens that restrict subagents to a specific "intent-scope" of a parent's permissions.

### Priority Shifts
- **Recursive Context Protocol**: Promoted from **P1** to **P0**. Essential for modern agent swarms to prevent state loss.
- **Shared KV Store**: Promoted from **P1** to **P0**. Critical for coordinating multi-agent actions in complex workflows.

### Deprecations / Monitoring
- *None today.*

---

## Evolution: [2026-02-24] Updates

### Proposed Additions
- **Advanced Multi-Agent Session Management**: (P0) A session-aware middleware that tracks tool state and handoffs between multiple specialized agents.
- **Unified MCP Discovery Service**: (P1) Automated discovery and registry for local and remote MCP servers (Stdio, HTTP, FastMCP).
- **Session-Bound State Persistence**: (P1) Ensuring that multi-agent "long-running" tasks maintain state across tool calls and agent switches.

### Priority Shifts
- **Policy Firewall**: Promoted to **P0** to support secure "Zero Trust" subagent isolation as ecosystems become more complex.

---

## Evolution: [2026-02-25] Updates

### Proposed Additions
- **On-Demand Discovery Middleware (Lazy-MCP)**: (P0) Implements similarity-based tool searching to prevent context pollution. Essential for massive (100+) tool libraries.
- **MCP Provenance Attestation**: (P1) Cryptographic verification of MCP server origins to prevent "Clinejection"-style supply chain attacks.
- **Slash-Command Bridge for Gemini**: (P1) Automatic mapping of MCP prompts to native Gemini CLI slash commands.

### Priority Shifts
- **Environment Bridging Middleware**: Promoted from **P1** to **P0**. The need for secure "Local-to-Cloud" tool bridging is increasing with more agents running in remote sandboxes.
- **Supply Chain Integrity Guard**: (New entry but P0 priority) High urgency due to recent ecosystem exploits.

### Deprecations / Monitoring
- **Upfront Tool Schema Pushing**: Monitoring for deprecation in favor of Lazy-Discovery.

---

## Evolution: [2026-02-26] Updates

### Proposed Additions
- **A2A Interop Bridge (Pseudo-MCP)**: (P0) Allows agents to interact with other agent frameworks using the A2A protocol, exposed as standard MCP tools.
- **Federated MCP Node Peering**: (P1) Secure discovery and proxying of tools across distributed MCP Any instances.
- **Cost & Latency Telemetry Middleware**: (P1) Automatically injects performance metadata into tool schemas to enable resource-aware agent reasoning.

### Priority Shifts
- **MCP Provenance Attestation**: Promoted to **P0** as it is a prerequisite for secure Federated MCP peering.
- **Lazy-MCP Middleware**: Promoted to **P0** (Already P0, but re-affirming importance for Federated Tool Mesh).

### Deprecations / Monitoring
- **Static Tool Schemas**: Moving towards dynamic, metadata-rich schemas that include real-time performance metrics.

---

## Evolution: [2026-02-28] Updates

### Proposed Additions
- **"Safe-by-Default" Network Hardening**: (P0) Transition to local-only default bindings for all services. Requires explicit MFA/Attestation for remote exposure.
- **A2A Stateful Residency (Stateful Buffer)**: (P0) MCP Any acts as a persistent mailbox for A2A messages, enabling reliable communication between agents with intermittent connectivity.
- **Provenance-First Discovery (Attested Discovery)**: (P1) Automatic filtering of MCP servers based on cryptographic signatures and community reputation scores.

---
---

## Evolution: [2026-03-09] Updates

### Proposed Additions
- **Project Configuration Security Guard**: (P0) Validating proxy for project-local agent configurations (e.g., `.claude/settings.json`) to prevent RCE via malicious hooks.
- **Agent-Aware Blackboard Isolation**: (P0) Implements row-level security for the Shared KV Store, ensuring agents can only access state within their assigned "Intent Scope."
- **Detached Sandbox for Automated Hooks**: (P1) Isolated execution environment for automated tool sequences, preventing unauthorized host access.

### Priority Shifts
- **Shared KV Store (Blackboard)**: Re-affirmed as **P0** with new mandatory security isolation requirements.
- **Policy Firewall**: Promoted to **P0** (Already P0, but expanded to include "Project-Local Config Validation").

### Deprecations / Monitoring
- **Unvalidated Project-Local Configs**: Monitoring for total deprecation. All local configs must be attested via MCP Any before ingestion by agents.

---

## Evolution: [2026-03-10] Updates

### Proposed Additions
- **Sandbox-as-a-Service for Config Hooks**: (P0) A natively managed, ultra-lightweight execution environment for approved hooks found in project-local settings.
- **Project Configuration Drift Detection**: (P1) Background monitor that alerts the user if a project-local configuration file is modified (e.g., via `git pull`), requiring re-attestation of any hooks.
- **Intent-Bound Context Isolation**: (P0) Cryptographic enforcement that prevents subagents from accessing state or tools outside their explicitly assigned "Intent-Scope."

### Priority Shifts
- **Detached Sandbox for Automated Hooks**: Promoted from **P1** to **P0**. Urgent requirement to mitigate RCE vulnerabilities discovered in the ecosystem.
- **A2A Interop Bridge**: Re-affirmed as **P0** to support secure state handoffs in multi-agent swarms.

### Deprecations / Monitoring
- **Implicit Hook Execution**: All "hooks" or "auto-exec" commands in configurations are now **Deprecated**. They must be explicitly moved to an "Attested Hooks" registry.

---

## Evolution: [2026-03-11] Updates

### Proposed Additions
- **Project-Local Config Attestation Engine**: (P0) A core service that intercepts and verifies cryptographic signatures on project-local configuration files.
- **Base-URL Hijack Protection (Exfiltration Guard)**: (P0) A middleware that enforces a strict "Allow-List" for LLM base URLs, preventing silent redirection of API traffic.
- **Active Config Rewriter**: (P1) A daemon that monitors agent configuration files and automatically reverts unauthorized changes to security-critical fields.

---

## Evolution: [2026-03-12] Updates

### Proposed Additions
- **Verified Skill Registry**: (P0) A security-first marketplace/registry for agent skills, requiring behavioral profiling and cryptographic signing before installation.
- **Offline-First Resilient Proxy**: (P1) A hardened gateway that handles complex proxy configurations and provides a stable LLM interface for air-gapped or restricted environments.
- **MFA for Project-Local Hooks**: (P0) Extends the HITL Middleware to require multi-factor attestation for any executable hook found in project configurations.

---

## Evolution: [2026-03-13] Updates

### Proposed Additions
- **OpenClaw ContextEngine Bridge**: (P1) A middleware that enables MCP Any to synchronize state with OpenClaw's new pluggable ContextEngine.
- **Prompt Path Protection Middleware**: (P0) Real-time scanning of tool outputs for "Indirect Prompt Injection" patterns to prevent agent hijacking.
- **Critical Skill Simulation (Dry-Run 2.0)**: (P1) Advanced "what-if" analysis for skills that simulates their impact on sensitive data before they are executed.
- **Swarm Behavioral Baseline**: (P1) Monitoring tool to establish a "normal" behavior pattern for agent swarms and alert on anomalies.

### Priority Shifts
- **Verified Skill Registry**: Re-affirmed as **P0** following the "ClawHavoc" malicious skill crisis.
- **A2A Interop Bridge**: Re-affirmed as **P0** to support the industry shift towards "Agentic Swarms."

### Deprecations / Monitoring
- **Direct Agent-to-LLM Communication**: Monitoring for deprecation in favor of **Exfiltration-Resistant Transport** (Proxied via MCP Any).
- **Unsigned/Unverified Skills**: Moving towards a default-block policy for any skill not present in the Verified Skill Registry.

---

## Evolution: [2026-03-14] Updates

### Proposed Additions
- **Same-Origin Policy (SOP) Enforcer for MCP**: (P0) Middleware that validates `Origin` and `Sec-Fetch-Site` headers for all local requests to prevent cross-site hijacking (CVE-2026-25253).
- **Context Lifecycle Hooks**: (P1) Pluggable lifecycle hooks for context creation, compression, and retrieval, enabling custom "Intent-Preserving" strategies.
- **Semantic Boundary Detector**: (P0) A specialized scanning module for the Prompt Path Protection middleware that detects malicious instructions hidden in multimodal metadata (SVG, CSS).
- **Session-Resumption mTLS for Swarms**: (P1) Optimized mTLS transport that uses session tickets to reduce handshake latency in high-frequency A2A communication.

### Priority Shifts
- **OpenClaw ContextEngine Bridge**: Promoted to **P0**. Urgent need for interoperability to combat "Context Ghosting" in shared swarms.
- **"Safe-by-Default" Network Hardening**: (Re-affirmed P0) Expanded to include mandatory browser-origin validation for all local listeners.

### Deprecations / Monitoring
- **Unvalidated Local WebSockets**: Monitoring for total deprecation. All local WebSocket connections must provide a valid, allow-listed `Origin` header.

---

## Evolution: [2026-03-15] Updates

### Proposed Additions
- **Call-Graph Loop Monitor**: (P0) Middleware to detect and prevent recursive "M2M" tool loops that cause resource exhaustion.
- **Signed Context Chain Protocol**: (P0) Cryptographic signing of subagent requests to prevent identity spoofing (CVE-2026-28190).
- **Universal Agent Bus (UAB) Adapter**: (P1) Native support for the UAB protocol, enabling seamless task handoffs between OpenClaw and AutoGen frameworks.

---

## Evolution: [2026-03-16] Updates

### Proposed Additions
- **Browser-Origin Validation Middleware**: (P0) Mandatory validation of `Origin` and `Sec-Fetch-Site` headers for all local listeners to mitigate cross-site hijacking (CVE-2026-25253).
- **UAB Task Delegation Bridge**: (P1) Extension of the A2A bridge to support UAB-native task cards and authenticated discovery.
- **Cross-Agent Loop Circuit Breaker**: (P0) Real-time monitoring of inter-agent call graphs to prevent "Spiral of Death" loops across framework boundaries.
- **Relational Identity Provider**: (P1) A core service that maps and verifies agent identities between disparate frameworks (e.g., OpenClaw, Gemini CLI).

### Priority Shifts
- **Signed Context Chain Protocol**: Re-affirmed as **P0** with expanded requirements for UAB compatibility.
- **"Safe-by-Default" Network Hardening**: (Re-affirmed P0) Now includes mandatory Browser-Origin enforcement for all adapters.

### Deprecations / Monitoring
- **Implicit Local Trust**: All listeners must now explicitly validate request origins. Standard `localhost` binding without header checks is now **Deprecated**.

---

## Evolution: [2026-03-17] Updates

### Proposed Additions
- **Local-Loopback Rate Limiter**: (P0) Mandatory rate limiting for all `127.0.0.1` and `::1` connections to prevent brute-force attacks on gateway credentials.
- **Behavioral Skill Burn-In Sandbox**: (P1) An isolated environment where new skills are profiled for "Delayed Payload" behaviors before being promoted to "Trusted" status.
- **UAB Authenticated Task Delegation Bridge**: (P0) Full implementation of UAB v1.2 "Authenticated Task Cards" for secure cross-framework delegation.
- **Local Security Audit Log**: (P1) Detailed logging of all local connection attempts, including origin headers and authentication success/failure rates.

### Priority Shifts
- **Universal Agent Bus (UAB) Adapter**: Promoted to **P0**. Essential for cross-framework agentic coordination.
- **Verified Skill Registry**: (Re-affirmed P0) Expanded to include Behavioral Profiling requirements.

### Deprecations / Monitoring
- **Unthrottled Local Access**: All local interfaces must now implement rate limiting. Unthrottled loopback access is now **Deprecated**.

---

## Evolution: [2026-03-18] Updates

### Proposed Additions
- **Local Listener Origin Enforcement**: (P0) Mandatory `Origin` and `Sec-Fetch-Site` validation for all local API/WebSocket listeners to prevent cross-site hijacking.
- **Recursive Depth-Limit Middleware**: (P0) Advanced call-graph monitoring to detect and block infinite tool-calling loops across different agents.
- **UAB Authenticated Task Delegation Core**: (P0) Full implementation of UAB task card verification, ensuring all cross-framework delegations are authenticated.
- **Lineage-Aware Context Signing**: (P1) Cryptographic signing of the entire context chain to prevent subagent identity spoofing.

---

## Evolution: [2026-03-19] Updates

### Proposed Additions
- **UACO-Native Coordination Middleware**: (P0) Full implementation of the Universal Agent Coordination Protocol for task negotiation, bidding, and stateful handoffs.
- **Unified RL Feedback Telemetry Bridge**: (P1) Middleware for collecting and normalizing agent performance and conversation feedback for RL training loops (e.g., OpenClaw-RL).
- **Enterprise Policy Sync Engine**: (P1) Core service for synchronizing security policies and allowed-origin lists from a centralized enterprise management server.

---

## Evolution: [2026-03-20] Updates

### Proposed Additions
- **Ephemeral Workspace Trust Middleware**: (P0) A session-bound attestation service that translates desktop-level trust tokens into persistent agent capabilities.
- **Blackboard Integrity Validator**: (P0) Cryptographic validation for all Shared KV Store operations, ensuring state lineage and intent-bound isolation.
- **UACO Bid Profiling Engine**: (P1) Behavioral monitoring service that evaluates agent bids against historical performance and safety baselines to prevent "Task Card Shadowing."
- **Config Smuggling Scanner**: (P1) Specialized scanner for project-local configurations that detects malicious instructions hidden in binary/metadata blobs.

### Priority Shifts
- **A2A Interop Bridge**: Promoted to **P0**. With UACO maturation, the bridge is now critical for multi-agent task negotiation.
- **Project Configuration Security Guard**: (Re-affirmed P0) Expanded to include support for Enterprise-Managed policy overrides.

### Deprecations / Monitoring
- **Framework-Specific Feedback Logs**: Monitoring for deprecation. Feedback should be normalized via the Unified Telemetry Bridge.

---

## Evolution: [2026-03-21] Updates

### Proposed Additions
- **Content-Addressable Config (CAC) Validator**: (P0) A core security service that enforces hash-based validation for all executable hooks and settings, preventing "Binary Smuggling."
- **UACO v1.5 RCC Validator**: (P0) Implementation of Resource Capability Claims to verify agent toolsets and security posture during task bidding.
- **DNS/ICMP Exfiltration Monitor**: (P1) L4 telemetry middleware to detect and block non-HTTP exfiltration attempts by compromised agents.
- **Hardware-Bound Trust Continuity**: (P1) Extension for the Ephemeral Workspace Trust Middleware that uses TPM/Secure Enclave signatures to persist trust for headless agents.

---

## Evolution: [2026-03-22] Updates

### Proposed Additions
- **UACO Agentic SLA Middleware**: (P0) Enforcement layer for resource contracts (token budget, reasoning time) during UACO task delegation.
- **Federated Policy Synchronizer**: (P1) A secure bus for synchronizing CAC hashes and allowed-origin lists across multiple MCP Any instances.
- **Ghost Shell Execution Mode**: (P0) Isolated, instrumented profiling environment for un-attested hooks, providing behavioral insights before attestation.

### Priority Shifts
- **UACO v1.5 RCC Validator**: Re-affirmed as **P0**. Essential foundation for the new SLA middleware.
- **Shared KV Store (Blackboard)**: (Re-affirmed P0) Expanded to support "SLA-Aware State Locking" to prevent resource-heavy contention.

### Deprecations / Monitoring
- **Unbounded Task Delegation**: Moving toward total deprecation. All UACO delegations must eventually include a resource contract (SLA).

---

## Evolution: [2026-03-23] Updates

### Proposed Additions
- **Proof-of-Intent (PoI) Validator**: (P0) Middleware that implements UACO v1.7 headers to verify that tool calls align with cryptographically signed session intents.
- **Binary State Handoff (BSH) Gateway**: (P1) High-performance binary transport for agent state to mitigate "Token Storms" and JSON overhead.
- **Multi-Signature Skill Attestation**: (P0) Verification mechanism for dynamic skill grafting, requiring signatures from both framework and user policy to prevent "Skill-Squatting."

### Priority Shifts
- **UACO-Native Coordination Middleware**: Re-affirmed as **P0**. Urgent update required to support v1.7 PoI and combat Context-Mirroring.
- **Verified Skill Registry**: (Re-affirmed P0) Expanded to include real-time attestation for dynamic grafting.

### Deprecations / Monitoring
- **JSON-only State Handoffs**: Monitoring for deprecation in favor of **BSH** for high-frequency agent swarms.

---

## Evolution: [2026-03-24] Updates

### Proposed Additions
- **Relational PoI Validator**: (P0) Extends PoI validation to verify the entire "Intent Chain," ensuring subagents cannot be coerced into actions outside the parent's verified goal.
- **BSH State Buffer**: (P1) High-speed memory-mapped buffer for binary state handoffs between agents to minimize context transfer latency.
- **Ghost Shell Hook Profiler**: (P0) Instrumented sandbox for behavioral profiling of un-attested configuration hooks, detecting "Binary Smuggling" before host execution.

### Priority Shifts
- **Binary State Handoff (BSH) Gateway**: Promoted from **P1** to **P0**. Urgent requirement to solve the "Token Storm" crisis in deep swarms.
- **Ghost Shell Execution Mode**: Re-affirmed as **P0**. Critical security defense against malicious project-local hooks.
