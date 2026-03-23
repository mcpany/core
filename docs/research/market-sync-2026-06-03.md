<!--
  Copyright 2026 Author(s) of MCP Any

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
-->

# Copyright 2026 MCP Any Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Market Sync: 2026-06-03
**Objective:** Evolution of Project-Local Policy Sovereignty and Source-Instruction Integrity.

## Ecosystem Shifts

### 1. OpenClaw v2026.3.1: "Advanced Agent Governance"
* **Observation:** OpenClaw has released a patch (v2026.3.1) addressing the "Identity-Budget Divergence" in parallel teams.
* **Technical Shift:** Introduction of `Task-Scoped Identity Propagation`. This allows subagents to inherit a "Hardware-Attested Task Token" that automatically expires when the specific task ID is marked as completed in the shared CRDT task-list.
* **Trend:** Shift from "Session-Bound" to "Task-Bound" identity lifecycles.

### 2. Gemini CLI v0.30.0: "Project-Level Policy Engine"
* **Observation:** Google has released Gemini CLI v0.30.0 with a native `gcloud-policy` adapter.
* **Technical Shift:** Gemini now natively ingests `.gemini/policy.cel` files for project-local tool-call gating. It uses Common Expression Language (CEL) to define "Safe Zones" for tool execution.
* **Trend:** Standardization of "Policy-as-Code" within the local project directory.

### 3. Claude Code: "Mesh-Bound Reasoning Verification"
* **Observation:** Anthropic is testing a beta feature for Claude Code that cross-references reasoning traces between parallel teammates.
* **Technical Shift:** A "Verifier Teammate" can now challenge the reasoning of a "Worker Teammate" before a tool call is authorized, effectively implementing a real-time, 2-agent quorum for all high-stakes shell commands.
* **Trend:** Transition from "Supervisor-led" to "Peer-verified" swarm integrity.

### 4. New Threat: "oh-my-opencode" Prompt Injection (CVE-2026-11822)
* **Observation:** A critical vulnerability has been disclosed in the popular `oh-my-opencode` agentic shell wrapper.
* **Threat Pattern:** Attackers are using "Source-Instruction Smuggling" where malicious instructions are hidden in `.gitignore` or `.env.example` files. Since agents often read these to build context, they unknowingly ingest "Hidden System Prompts" that override the user's mission.
* **Requirement:** Mandatory "Source-Instruction Integrity Guard" (SIIG) that sanitizes all project-local metadata files before they enter the agent context.

## Unique Findings for Today

* **The Policy Sovereignty Gap:** While Gemini CLI has its own policy engine and OpenClaw has another, there is no "Universal Policy Adapter" that can reconcile `.cel` and `.rego` policies across a heterogeneous swarm. MCP Any is perfectly positioned to be this **Universal Policy Gate**.
* **Source-Instruction Integrity:** The "oh-my-opencode" exploit proves that context-pollution is no longer just about external web results; the project's own metadata is now a weaponized input.

## Strategic Impact

1. **Project-Level Policy Engine Adapter:** MCP Any must evolve to ingest and reconcile disparate project-local policies (Gemini's CEL, OpenClaw's Rego) into a single, hardware-attested gate.
2. **Source-Instruction Integrity Guard (SIIG):** We should implement a specialized scanner for project metadata files (`.gitignore`, `.env`, `.github/`) to detect smuggled system-prompt overrides.
3. **Hardware-Attested Policy Discovery:** Policy files themselves must be hardware-attested to prevent a malicious subagent from "Settings-Squatting" and rewriting the project's security rules.
