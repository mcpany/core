# Design Doc: Runtime Egress Watcher (AgentMoat)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents move toward autonomy, the primary risk shift is from "What can they say?" to "What can they do?". Current MCP Any security relies on static permissions (e.g., "allow read on /tmp"). However, a compromised agent (via Indirect Prompt Injection) could use those permitted tools to execute malicious actions that still fall within the static scope.

AgentMoat introduces runtime behavioral monitoring at the host level, ensuring that any action initiated by an MCP Any tool is monitored, logged, and compared against the agent's stated intent. It bridges the gap between high-level LLM instructions and low-level system calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Real-time monitoring of filesystem, network, and process activity triggered by tools.
    * Correlating low-level activity with the high-level Tool Call ID and Agent Session.
    * Providing "Forbidden Zones" and "Egress Allow-lists" that tools cannot bypass.
    * Integration with NIST-aligned runtime reporting requirements.
* **Non-Goals:**
    * Providing a full sandbox (e.g., gVisor/Docker replacement). AgentMoat *monitors* and *restricts* existing tool execution environments.
    * Replacing existing Model-level guardrails (e.g., Llama Guard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise DevOps
* **Primary Goal:** Prevent an autonomous agent from exfiltrating sensitive `.env` files even if it has general "File Read" permissions.
* **The Happy Path (Tasks):**
    1. Admin configures an AgentMoat policy in MCP Any defining "Sensitive Data" paths (e.g., `**/.env`).
    2. An agent is tasked with "Reviewing project structure."
    3. The agent calls the `read_file` tool on a `.env` file (indirectly injected).
    4. AgentMoat detects the system-level open() call on the sensitive path.
    5. AgentMoat cross-references the call with the tool's intent. Since "Reviewing project structure" doesn't require `.env` access, it blocks the call.
    6. The block is logged with full context (Session ID, Intent, Tool Call).

## 4. Design & Architecture
* **System Flow:**
    * MCP Any (Gateway) -> Tool Execution Middleware -> AgentMoat (Runtime Monitor) -> System Call.
    * AgentMoat uses eBPF (on Linux) or host-level hooks to monitor activity.
    * Data Flow:
        1. Tool call is received with an `intent_metadata` header.
        2. Middleware registers the `Thread/Process ID` with AgentMoat.
        3. AgentMoat enforces policies on that specific PID for the duration of the tool call.
* **APIs / Interfaces:**
    * `POST /v1/policies/egress`: Define new monitoring rules.
    * `GET /v1/monitoring/alerts`: Stream real-time behavioral alerts.
* **Data Storage/State:**
    * Behavioral logs stored in a high-performance append-only SQLite log (audit-db).

## 5. Alternatives Considered
* **Docker-only Isolation:** Rejected because containerization doesn't provide fine-grained intent-awareness. An agent in a container can still exfiltrate data *within* that container.
* **AppArmor/SELinux profiles:** Rejected as too static and difficult for LLM-based systems to manage dynamically.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AgentMoat itself must run with high privileges (to monitor syscalls) but must expose a minimal, authenticated API to MCP Any.
* **Observability:** Performance impact must be <5% overhead on tool execution. eBPF is chosen for its efficiency.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
