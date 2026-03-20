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

# Design Doc: Supply Chain Provenance Attestor (SCPA)

**Status:** Draft
**Created:** 2026-06-05

## 1. Context and Scope

The rapid expansion of the agentic ecosystem has led to a critical reliance on third-party tool registries and dynamic library updates. Recent security reports (Barracuda 2026) have identified over 43 agent framework components with backdoors introduced via supply chain compromise, often remaining dormant until activated by a remote command-and-control server. The "Clinejection" attack pattern demonstrates how unverified tool definitions and library updates can be weaponized to exfiltrate credentials and PII.
MCP Any must evolve from a passive tool proxy to an active guarantor of supply chain integrity. The SCPA provides a hardware-bound mechanism for attesting to the lineage and structural integrity of every tool and library in the agent's execution path. By anchoring provenance in a Trusted Platform Module (TPM), we ensure that tool definitions cannot be tampered with or replaced by malicious updates without a verified cryptographic signature.

## 2. Goals & Non-Goals

*   **Goals:**

    *   Provide hardware-attested (TPM/Secure Enclave) signatures for all tool definitions and local libraries.

    *   Enforce mandatory signature validation for all tool updates before they are committed to the active environment.

    *   Facilitate multi-signature quorums for high-risk upstream tool updates.

    *   Ensure that tool metadata (schemas, descriptions) remains immutable once attested.

*   **Non-Goals:**

    *   Providing general-purpose code signing for non-agentic host applications.

    *   Verifying the functional "safety" or correctness of the tool logic (this is the responsibility of the behavioral profiling and Ghost Shell layers).

## 3. Critical User Journey (CUJ)

*   **User Persona:** Enterprise AI Infrastructure Architect

*   **Primary Goal:** Prevent a poisoned "Database Specialist" tool update from compromising a high-clearance swarm.

*   **The Happy Path (Tasks):**

    1.  An update for the "Database Specialist" tool manifest is received by the local MCP Any registry.

    2.  SCPA intercepts the update and issues a TPM-bound signature challenge to the originating developer workstation.

    3.  The developer signs the update using a hardware-resident key; SCPA verifies this against the local "Provenance Root."

    4.  A multi-agent security quorum (using the APRIG middleware) performs a pre-commit semantic scan of the new tool schemas.

    5.  SCPA issues a unique "Provenance Token" and allows the update to be merged into the active capability card.

    6.  The agent swarm boots, validating the Provenance Token before establishing any tool sessions.

## 4. Design & Architecture

*   **System Flow:**

    `Registry Update Request -> SCPA Interceptor -> Hardware Signature Validation -> APRIG Multi-Agent Review -> Provenance Token Issuance -> Capability Mapping`

*   **APIs / Interfaces:**

    *   `/v1/provenance/sign`: Endpoint for submitting a tool manifest for hardware signing.

    *   `/v1/provenance/verify`: Internal service API for validating tokens during agent boot.

    *   `mcp.provenance.v1`: A gRPC-based extension for the PNTD provider to carry provenance metadata.

*   **Data Storage/State:**

    *   **Provenance Root Keys**: Stored in the local hardware TPM or Secure Enclave.

    *   **Provenance Ledger**: A versioned, append-only ledger stored on the Blackboard that tracks the signature history and lineage of every attested tool.

## 5. Alternatives Considered

*   **Centralized SaaS Attestation**: Rejected because it introduces a single point of failure and violates the "Local Sovereignty" mandate for air-gapped and high-security enterprise environments.

*   **Plain SHA-256 File Hashing**: Rejected because hashes only prove integrity, not provenance. An attacker with filesystem access could replace both the tool and the expected hash simultaneously. SCPA's use of hardware-bound asymmetric keys prevents this "Shadow Substitution."

## 6. Cross-Cutting Concerns

*   **Security (Zero Trust):** SCPA implements "Provenance-before-Discovery." An agent cannot even "see" a tool in the discovery bus unless it carries a valid, hardware-attested provenance token. This prevents "Shadow Tool" injection during the mesh handshake.

*   **Observability:** Every signature validation and provenance failure is logged to the Local Security Audit Dashboard. The UI provides a "Provenance Heatmap" showing the trust strength of the entire tool supply chain.

## 7. Evolutionary Changelog

*   **2026-06-05:** Initial Document Creation.
