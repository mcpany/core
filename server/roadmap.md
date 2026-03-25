<!-- markdownlint-disable -->
# Server Roadmap

## 1. Top Priorities: The Universal Agent Bus (New Strategic Focus)
*   **[Security] Policy Firewall Engine:** Implement Rego/CEL based hooking for tool calls.
*   **[Security] Granular Scopes:** implement capability-based token system (`fs:read:/tmp`).
*   **[P0] HAIL Lineage Provider (SRV-2026-06-19):** Authoritative security middleware that issues cryptographically signed "Reasoning Fragments" for all inter-agent sub-instructions.
*   **[P0] HLAG Attention Governor (SRV-2026-06-19):** Advanced attention guard utilizing hardware-bound headers to "pin" mission-critical fragments.

## 2. Shared Context & Memory (Sovereignty-First)
*   **[P0] LFMC Hub (SRV-2026-06-19):** Coordination service implementing CRDT-based task list synchronization for parallel teammates.
*   **[P1] SCI Metadata Interceptor (SRV-2026-06-18):** Advanced security for the T2T Bridge that monitors out-of-band collusion.
*   **[Context Engine] Sharded SQLite Middleware:** Implement granular context shards as distinct database files for hardware-bound isolation.
*   **[Context Engine] Hierarchical TTL:** Support TTL policies on context shards based on sub-mission completion.
