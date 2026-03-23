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

# Design Doc: Project-Level Policy Gate
**Status:** Draft
**Created:** 2026-06-03

## 1. Context and Scope
With the release of Gemini CLI v0.30.0 and OpenClaw v2026.3.1, the ecosystem is fragmenting into disparate, framework-specific project-local policy formats (CEL and Rego). Currently, a multi-agent swarm using both Gemini specialists and OpenClaw teammates cannot enforce a consistent security boundary because each agent only respects its own local config. MCP Any must act as the "Universal Policy Gate" that reconciles these project-local files into a single, hardware-attested manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Ingest `.gemini/policy.cel` and `.openclaw/policy.rego` from the local project root.
    * Reconcile conflicting rules into a "Most Restrictive" hardware-attested manifest.
    * Enforce this reconciled policy at the tool-call middleware layer for all connected agents.
    * Provide a unified "Policy Violation" event stream to the UI.
* **Non-Goals:**
    * Replacing the underlying CEL/Rego engines (we wrap them).
    * Synchronizing local project policies to the cloud (they stay local for sovereignty).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that all agents in a heterogeneous swarm respect the project's "No-Outbound-Internet" rule, regardless of whether they are Gemini or OpenClaw based.
* **The Happy Path (Tasks):**
    1. User adds a `policy.cel` file to the project root with a tool-call restriction.
    2. MCP Any detects the file via the Hardware-Attested Policy Discovery Provider.
    3. The Policy Gate reconciles the CEL rule with existing mesh-wide Rego policies.
    4. A subagent (OpenClaw) attempts to call a `fetch` tool.
    5. The Policy Gate intercepts the call, evaluates it against the reconciled manifest, and blocks it.
    6. The UI surfaces a "Project-Policy Violation" alert with a signed justification from the TPM.

## 4. Design & Architecture
* **System Flow:**
    `[Local Project Files] -> (TPM Sign) -> [Policy Discovery] -> [Reconciliation Engine] -> (Manifest) -> [Tool Call Middleware]`
* **APIs / Interfaces:**
    * `policy.v1.GetReconciledManifest()`: Returns the currently active, signed policy.
    * `policy.v1.ValidateAction(action_request)`: Middleware hook for per-call validation.
* **Data Storage/State:**
    * In-memory cache of the reconciled CEL/Rego state, refreshed only upon hardware-attested file change events.

## 5. Alternatives Considered
* **Framework-Specific Gateways:** Rejected because it forces the user to maintain duplicate policies and doesn't solve the "Shadow Agent" problem where one agent ignores another's policy.
* **Manual Conversion:** Rejected as it's too slow and prone to human error in machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All policy files must be hardware-attested (TPM) to prevent a compromised agent from rewriting its own constraints (neutralizing "Settings-Squatting").
* **Observability:** Reconciled policy state is visible in the "Governance Center" UI.

## 7. Evolutionary Changelog
* **2026-06-03:** Initial Document Creation.
