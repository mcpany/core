# Design Doc: DHT-Native Discovery Proxy
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The shift toward multi-device, heterogeneous agent meshes (e.g., OpenClaw v3.7) has exposed the limitations of centralized capability discovery. Registry bottlenecks and single-point-of-failure risks in local network coordination often lead to "Discovery Latency" that degrades the agentic experience.

MCP Any must evolve from a centralized registry to a Distributed discovery model. The DHT-Native Discovery Proxy enables MCP Any to act as a high-speed node in a Distributed Hash Table (DHT), allowing agents to discover and invoke tools across multiple local and remote devices with sub-millisecond latency and hardware-attested security.

## 2. Goals & Non-Goals
* **Goals:**
    * Support OpenClaw DCD (Distributed Capability Discovery) standard.
    * Implement sub-millisecond capability lookups across multi-device meshes.
    * Enforce hardware-attested "Auth-before-Discovery" at the DHT layer.
    * Neutralize "Discovery Flooding" via hardware-locked rate limits on beacons.
* **Non-Goals:**
    * This proxy WILL NOT manage the execution of remote tools (handled by AMT Broker).
    * This proxy WILL NOT store raw tool schemas in the DHT (only capability hashes and node pointers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cross-Device Agent User (Desktop, Phone, and Raspberry Pi cluster)
* **Primary Goal:** A phone-based agent discovers a high-power GPU tool on the desktop node without a central server.
* **The Happy Path (Tasks):**
    1. Desktop Node publishes a "Capability Hash" and its AMT endpoint to the DHT.
    2. Phone Agent broadcasts a "Capability Query" signed with its hardware mission-token.
    3. The DHT-Native Discovery Proxy on the desktop node receives the query and validates the mission-token.
    4. The Proxy returns the encrypted tool schema and connection pointer to the Phone Agent.
    5. Phone Agent initiates a secure tunnel via the AMT Broker to execute the tool.

## 4. Design & Architecture
* **System Flow:**
    * [Capability Provider] -> Register(Hash) -> Local DHT Proxy -> [DHT Mesh]
    * [Agent] -> Search(Capability) -> Local DHT Proxy -> [DHT Lookup] -> [Node Pointer]
* **APIs / Interfaces:**
    * `discovery.v1.DHTService/Publish`: Broadcasts local capabilities to the mesh.
    * `discovery.v1.DHTService/Lookup`: Queries the mesh for a specific capability hash.
* **Data Storage/State:**
    * Kademlia-based routing table for peer nodes.
    * Hardware-attested cache of capability pointers.

## 5. Alternatives Considered
* **Centralized Discovery Hub**: Rejected as it creates a bottleneck and a single point of failure in large meshes.
* **mDNS/Bonjour only**: Rejected because it lacks the routing efficiency and hardware-attested security required for cross-device agent coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * "Auth-before-Discovery" (ZKD Proxy compliant) ensures capability details remain masked from unauthenticated peers.
    * Every DHT join and query requires a TPM-signed node identity and mission-token.
* **Observability:**
    * Real-time mesh topology is visualized in the UI (Service Mesh Topology Monitor).
    * Discovery latency and beacon volume metrics are tracked to detect flooding attacks.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
