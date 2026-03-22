# Design Doc: Federated A2A Registry (FAR)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As the Agent2Agent (A2A) ecosystem moves toward multi-organizational collaboration, the need for a decentralized and secure discovery mechanism is critical. Current discovery often relies on localized or unverified registries, making the mesh vulnerable to "Registry Squatting," where malicious agents intercept task delegations by registering shadowed identities.

FAR evolves MCP Any into a discovery node for the federated A2A mesh. It allows agents to publish and discover cryptographically signed "Agent Cards" across organization boundaries, ensuring that every peer in the mesh is verified and authorized for the mission scope.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a decentralized registry for A2A "Agent Cards."
    * Support cryptographic signature verification for agent identities.
    * Provide organization-level discovery constraints (e.g., "Only discover agents from Org X").
    * Integrate with the A2A v2.0 GA standard for federated discovery.
* **Non-Goals:**
    * Replacing the underlying A2A messaging transport.
    * Managing agent-local tool execution (handled by MCP/gRPC adapters).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cross-Enterprise Swarm Architect
* **Primary Goal:** Securely delegate a data-analysis task to a specialized agent hosted in a partner organization's infrastructure.
* **The Happy Path (Tasks):**
    1. The architect configures MCP Any as a FAR node.
    2. The partner organization publishes its "Analyst Agent Card" to the federated registry, signed by its corporate root CA.
    3. The local orchestrator agent queries FAR for "Analyst" capabilities within the partner's namespace.
    4. FAR returns the verified Agent Card and its endpoint.
    5. The orchestrator initiates an authenticated A2A handshake and delegates the task.

## 4. Design & Architecture
* **System Flow:**
    [Agent Request] -> [FAR Node] -> (Query Federation) -> [External FAR Nodes] -> (Return Verified Cards) -> [Agent Selection]
* **APIs / Interfaces:**
    * `POST /far/registry/publish`: Publishes a signed Agent Card.
    * `GET /far/registry/search`: Searches for agents across the federated mesh.
    * `POST /far/registry/verify`: Validates the signature and lineage of an external Agent Card.
* **Data Storage/State:**
    * Verified Agent Cards are cached in the Service Registry and backed by the Shared KV Store.

## 5. Alternatives Considered
* **Centralized Registry:** Rejected due to single-point-of-failure risks and organizational sovereignty concerns.
* **DHT-based Discovery:** Rejected because it lacks the strong organizational identity and attribution required for enterprise Zero-Trust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** FAR enforces "Auth-before-Discovery." Cards are only visible to authenticated peers with matching mission-root authority.
* **Observability:** Discovery events and verification failures are logged to the "Federated A2A Discovery Monitor."

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
