# Server Roadmap

## 1. Top Priorities: The Universal Agent Bus (New Strategic Focus)
*   **[Security] Policy Firewall Engine:** Implement Rego/CEL based hooking for tool calls.
*   **[Security] Granular Scopes:** implement capability-based token system (`fs:read:/tmp`).
*   **[Comms] Recursive Context Protocol:** Standardize headers for Subagent inheritance.
*   **[State] Shared Key-Value Store:** Embedded SQLite "Blackboard" tool for agents.
*   **[Security] HITL Middleware:** Suspension protocol for user approval flows.

## 2. Updated Roadmap

### Status: Active Development

#### Upcoming (2026-02-23 Evolution)
*   **[P0] Recursive Context Protocol**: Finalize header-based context inheritance for swarms.
*   **[P0] Zero-Trust Subagent Scoping**: Implement intent-bound capability tokens.
*   **[P1] Environment Bridging Middleware**: Secure state sync between cloud sandboxes and local tools.
*   **[P1] Machine-Checkable Security Contracts**: Declarative tool safety models.
*   **[P0] Multi-Agent Session Management**: Session-aware middleware for agent coordination (Added: 2026-02-24).
*   **[P1] Unified MCP Discovery Service**: Automated registry for Stdio/HTTP/FastMCP servers (Added: 2026-02-24).

#### Upcoming (2026-02-25 Evolution)
*   **[P0] On-Demand Discovery Middleware (Lazy-MCP)**: Implements similarity-based tool searching to prevent context pollution. (Added: 2026-02-25)
*   **[P0] Supply Chain Integrity Guard**: Cryptographic provenance verification for MCP servers to prevent unauthorized tool injection. (Added: 2026-02-25)
*   **[P1] FastMCP Metadata Support**: Support for Pythonic FastMCP decorators and native Gemini CLI slash command mapping. (Added: 2026-02-25)

#### Upcoming (2026-02-26 Evolution)
*   **[P0] A2A Interop Bridge**: Implement Pseudo-MCP wrapper for A2A-compliant agents. (Added: 2026-02-26)
*   **[P1] Federated MCP Peering**: Distributed node discovery and tool proxying. (Added: 2026-02-26)
*   **[P1] Resource Telemetry Middleware**: Inject latency/cost metrics into tool schemas. (Added: 2026-02-26)

#### Upcoming (2026-02-28 Evolution)
*   **[P0] Safe-by-Default Hardening**: Transition all listeners to `localhost` by default. Implement mandatory Attestation for remote exposure. (Added: 2026-02-28)
*   **[P0] A2A Stateful Residency**: Resident state for A2A messages, enabling asynchronous, reliable multi-agent handoffs. (Added: 2026-02-28)
*   **[P1] Provenance-First Discovery**: Cryptographic signature verification during tool discovery. (Added: 2026-02-28)

#### Upcoming (2026-03-10 Evolution)
*   **[P0] Sandbox-as-a-Service for Config Hooks**: Natively managed, ultra-lightweight execution environment for approved hooks found in project-local settings. (Added: 2026-03-10)
*   **[P0] Intent-Bound Context Isolation**: Cryptographic enforcement that prevents subagents from accessing state or tools outside their explicitly assigned "Intent-Scope." (Added: 2026-03-10)
*   **[P1] Project Configuration Drift Detection**: Background monitor that alerts the user if a project-local configuration file is modified (e.g., via `git pull`), requiring re-attestation of any hooks. (Added: 2026-03-10)

#### Upcoming (2026-03-09 Evolution)
*   **[P0] Project Configuration Security Guard**: Validating proxy for project-local agent configs (e.g., `.claude/settings.json`) to prevent RCE. (Added: 2026-03-09)
*   **[P0] Agent-Aware Blackboard Isolation**: Row-level security for Shared KV Store to prevent cross-agent state injection. (Added: 2026-03-09)
*   **[P0] Detached Sandbox for Automated Hooks**: Isolated runtime for tool sequences with zero host access by default. (Added: 2026-03-09 - Promoted to P0 on 2026-03-10)

#### Upcoming (2026-03-11 Evolution)
*   **[P0] Exfiltration-Resistant Transport Gateway**: Force all agent traffic through a secure, allow-listed proxy to prevent API key exfiltration. (Added: 2026-03-11)
*   **[P0] Project-Local Config Attestation Engine**: Cryptographic verification of signatures on agent configuration files. (Added: 2026-03-11)
*   **[P1] Active Config Rewriter**: Daemon that automatically reverts unauthorized changes to security-critical agent settings. (Added: 2026-03-11)

#### Upcoming (2026-03-12 Evolution)
*   **[P0] Verified Skill Registry**: Security-first marketplace/registry for agent skills requiring behavioral profiling. (Added: 2026-03-12)
*   **[P0] Mandatory MFA for Hooks**: Integration of HITL Middleware for multi-factor attestation of executable configuration hooks. (Added: 2026-03-12)
*   **[P1] Offline-First Resilient Proxy**: Hardened gateway for complex proxy configurations and air-gapped environment support. (Added: 2026-03-12)

#### Upcoming (2026-03-13 Evolution)
*   **[P0] Prompt Path Protection Middleware**: Real-time scanning of tool outputs for "Indirect Prompt Injection" patterns. (Added: 2026-03-13)
*   **[P0] OpenClaw ContextEngine Bridge**: Middleware to synchronize state with OpenClaw's pluggable context management. (Added: 2026-03-13 - Promoted to P0 on 2026-03-14)
*   **[P1] Critical Skill Simulation**: Advanced "what-if" analysis for skills, simulating impact on sensitive data. (Added: 2026-03-13)

#### Upcoming (2026-03-14 Evolution)
*   **[P0] Same-Origin Policy (SOP) Enforcer**: Middleware to validate browser-origin headers for local listeners (CVE-2026-25253). (Added: 2026-03-14)
*   **[P0] Semantic Boundary Detector**: Specialized scanner for Prompt Path Protection that analyzes multimodal metadata (SVG/CSS). (Added: 2026-03-14)
*   **[P1] Context Lifecycle Hooks**: Standardized API for framework-specific context compression and retrieval. (Added: 2026-03-14)
*   **[P1] Session-Resumption mTLS**: Optimized transport layer to reduce A2A handshake latency in large swarms. (Added: 2026-03-14)
*   **[P0] Authenticated A2A Agent Card Discovery**: Support for Gemini-style authenticated discovery in the A2A bridge. (Added: 2026-03-14)

#### Upcoming (2026-03-15 Evolution)
*   **[P0] Call-Graph Loop Monitor**: Middleware to detect and prevent recursive "M2M" tool loops and resource exhaustion. (Added: 2026-03-15)
*   **[P0] Signed Context Chain Protocol**: Cryptographic verification of subagent lineage to prevent identity spoofing (CVE-2026-28190). (Added: 2026-03-15)
*   **[P1] Universal Agent Bus (UAB) Adapter**: Native transport support for the UAB protocol for framework-neutral handoffs. (Added: 2026-03-15)

#### Upcoming (2026-03-16 Evolution)
*   **[P0] Browser-Origin Validation Middleware**: Mandatory validation of `Origin` and `Sec-Fetch-Site` headers for all local listeners (CVE-2026-25253). (Added: 2026-03-16)
*   **[P0] Cross-Agent Loop Circuit Breaker**: Real-time monitoring of inter-agent call graphs across framework boundaries. (Added: 2026-03-16)
*   **[P1] Relational Identity Provider**: Service to map and verify agent identities across disparate frameworks for context continuity. (Added: 2026-03-16)
*   **[P1] UAB Task Delegation Bridge**: Support for UAB-native task cards and authenticated discovery in the A2A bridge. (Added: 2026-03-16)

#### Upcoming (2026-03-17 Evolution)
*   **[P0] Local-Loopback Rate Limiter**: Mandatory rate limiting and auditing for all `127.0.0.1` / `::1` traffic to mitigate brute-force attacks. (Added: 2026-03-17)
*   **[P0] UAB Authenticated Task Delegation**: Core implementation of UAB v1.2 "Authenticated Task Cards" for secure cross-framework handoffs. (Added: 2026-03-17)
*   **[P1] Behavioral Skill Burn-In Sandbox**: Isolated profiling environment for detecting "Delayed Payload" malicious skills. (Added: 2026-03-17)
*   **[P1] Local Security Audit Service**: Background service for logging and analyzing local connection attempt patterns. (Added: 2026-03-17)

#### Upcoming (2026-03-19 Evolution)
*   **[P0] Active Intent Alignment (AIA) Broker**: Authoritative alignment service issuing hardware-attested heartbeats. (Added: 2026-03-19)
*   **[P0] UACO-Native Negotiation Hub**: Native implementation of UACO task bidding and stateful handoffs. (Added: 2026-03-19)
*   **[P0] Multi-Modal Behavioral Attestation (MMBA)**: Advanced identity service anchoring profiles to multi-modal history. (Added: 2026-03-19)
*   **[P1] Semantic State-Pinning (SSP)**: Attention governance middleware for mission-root fragment protection. (Added: 2026-03-19)
*   **[P0] Temporal Shard Jitter (TSJ) Injector**: Security extension for the ESB to neutralize timing side-channels. (Added: 2026-03-19)
