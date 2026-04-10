# Design Doc: Recursive Tunnel Attestation (RTA) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents increasingly operate across distributed physical nodes (OpenClaw SNT), the "Implicit Local Trust" assumption has become a primary attack vector. If an intermediate node in a P2P tunnel is compromised, it can inject malicious instructions into the tool-call stream. MCP Any needs to provide a mechanism to verify the complete hardware-attested lineage of every cross-node interaction.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate nested hardware signatures (TPM/SEP) for every hop in a multi-node tunnel.
    * Provide real-time verification of the "Chain of Sovereignty" for remote tool calls.
    * Neutralize "Intermediate-Node Injection" attacks.
* **Non-Goals:**
    * Replacing the underlying P2P transport (e.g., WireGuard, Tailscale).
    * Providing general-purpose network encryption.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed LLM Swarm Operator
* **Primary Goal:** Securely invoke a tool on a remote Raspberry Pi via an intermediate Bastion node without exposing the bastion to the intent.
* **The Happy Path (Tasks):**
    1. Agent initiates a tool call targeted at `Node-C`.
    2. `Node-A` (Source) signs the intent and encapsulates it in an RTA envelope.
    3. `Node-B` (Bastion) receives the envelope, adds its hardware-attested "Transit Signature," and forwards it.
    4. `Node-C` (Target) receives the multi-signed envelope.
    5. MCP Any RTA Broker on `Node-C` validates the nested signatures against the Mission Root.
    6. Tool is executed only if the lineage is verified.

## 4. Design & Architecture
* **System Flow:**
    `[Source Agent] -> [RTA Envelope A] -> [Transit Node] -> [RTA Envelope A+B] -> [Target Node (RTA Broker)]`
* **APIs / Interfaces:**
    * `POST /rta/sign`: Generate a transit signature for an existing envelope.
    * `POST /rta/verify`: Validate a nested signature chain against a mission manifest.
* **Data Storage/State:**
    * Short-lived "Transit Tickets" stored in kernel-bound memory.

## 5. Alternatives Considered
* **End-to-End Encryption Only**: Rejected because it doesn't prevent an intermediate node from dropping or replaying valid messages without a transit signature.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses TPM-bound monotonic counters to prevent replay of transit signatures.
* **Observability:** Visualized in the "Service Mesh Topology Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
