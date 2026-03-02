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

# Design Doc: Safe-by-Default Infrastructure Hardening

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The February 2026 security crisis (8,000+ exposed MCP servers, Clawdbot breach) highlighted a critical failure in the agentic ecosystem: ease-of-use was prioritized over security. Many users unknowingly bind MCP gateways to `0.0.0.0`, exposing sensitive tools and environment variables to the public internet. MCP Any must transition to a "Safe-by-Default" posture where the system is inherently secure even for novice users.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce `localhost` (`127.0.0.1` / `::1`) bindings for all adapters and gateways by default.
    *   Implement a "Remote Access Guard" that prevents `0.0.0.0` or non-local bindings without explicit administrative attestation.
    *   Introduce cryptographic MFA/Attestation for any remote management or tool access.
    *   Provide automated "Exposure Check" on startup.
*   **Non-Goals:**
    *   Completely disabling remote access (it must remain an option for enterprise use).
    *   Managing host-level firewall rules (MCP Any should focus on its own listener configuration).

## 3. Critical User Journey (CUJ)
*   **User Persona:** New AI Engineer deploying MCP Any for the first time.
*   **Primary Goal:** Set up the gateway without accidentally exposing tools to the internet.
*   **The Happy Path (Tasks):**
    1.  User runs `mcpany start` without a configuration file.
    2.  MCP Any binds all services to `127.0.0.1` and outputs a "Secure Local Mode" banner.
    3.  If the user attempts to set `host: 0.0.0.0` in the config, the server fails to start with a "Security Override Required" error.
    4.  User follows instructions to generate an `access_attestation.token` to enable remote exposure.

## 4. Design & Architecture
*   **System Flow:**
    - **Listener Configuration**: The `ConfigLoader` validates the `host` parameter. If non-local, it checks for a valid `AttestationToken`.
    - **Security Bootstrap**: On first run, a unique cryptographic identity (Ed25519) is generated for the instance.
    - **MFA Layer**: Remote access requests must include a signature from the instance's private key, typically handled via a "Second Screen" approval on the local machine.
    - **Anti-Vibe Constraint Engine**: Explicitly prevents "sudo" or "admin" level capabilities from being granted to any agent session by default. These are "High-Stakes" capabilities that require one-time cryptographic attestation per session.
*   **APIs / Interfaces:**
    - New CLI command: `mcpany secure authorize-remote [ip]`
    - New CLI command: `mcpany secure attest-high-stakes [session_id] [capability]`
    - Metadata extension for tool calls: `_mcp_source_locality: "local" | "remote"`
*   **Data Storage/State:** Secure storage of the instance identity in a protected file (e.g., `~/.mcpany/id_ed25519`).

## 5. Alternatives Considered
*   **Just Adding a Warning**: Log a warning when binding to `0.0.0.0`. *Rejected* as history shows users ignore logs.
*   **Requiring Docker Networking**: Forcing users to use Docker to isolate ports. *Rejected* as it adds too much friction for non-Docker workflows.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature is the foundation of the Zero Trust architecture. It ensures that the first point of failure (misconfiguration) is mitigated.
*   **Observability:** The UI should prominently display "Connectivity Status: [Local Only | Remote Authorized]" with a list of active remote listeners.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.

### Update: 2026-03-02 - Mitigating the "Vibe-Code" Vulnerability
**Context:** Recent market sync revealed a growing trend where developers grant over-privileged "sudo" access to agents because it "vibes" (works faster/easier), bypassing critical system safety.
**Architecture Adjustment:**
*   Introducing the **Anti-Vibe Constraint Engine** in Section 4.
*   Added `mcpany secure attest-high-stakes` CLI command to require explicit, one-time attestation for dangerous capabilities.
**Security Impact:** Prevents accidental or lazy exposure of host-level administrative access to rogue or hallucinating agents.
