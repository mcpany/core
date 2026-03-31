# Design Doc: Hardware-Enforced Loopback Isolation (HELI)
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) highlighted a critical failure in software-based origin validation for local loopback traffic. Malicious browser scripts can often bypass `Origin` header checks via socket-level or protocol-level exploits.

HELI provides a kernel-resident solution by utilizing eBPF filters to strictly isolate loopback traffic at the hardware/socket level. MCP Any needs this to ensure that only authorized local applications—not compromised browsers—can command the gateway.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement kernel-level eBPF filtering for all loopback (`127.0.0.1`, `::1`) traffic.
    * Mandate hardware-attested session handshakes for all socket connections.
    * Provide sub-millisecond interdiction of unauthenticated local probes.
* **Non-Goals:**
    * Replacing TLS for remote traffic.
    * Managing non-loopback network interfaces.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Local Developer
* **Primary Goal:** Prevent malicious websites from hijacking the local MCP Any control plane.
* **The Happy Path (Tasks):**
    1. User starts MCP Any with HELI enabled.
    2. MCP Any loads eBPF filter into the local kernel.
    3. Authorized local CLI tool attempts to connect.
    4. HELI filter validates the process-id and hardware-attested token.
    5. Connection is allowed.
    6. Malicious browser script attempts to open a WebSocket.
    7. HELI filter detects unauthorized origin/process and drops the packet at the kernel level.

## 4. Design & Architecture
* **System Flow:**
    `Process (CLI/Browser) -> Kernel (eBPF HELI Filter) -> MCP Any Gateway`
* **APIs / Interfaces:**
    * `HELI_Isolation_Init(policy_map)`: Initializes eBPF maps.
    * `HELI_Validate_Socket(fd)`: Callback for hardware attestation check.
* **Data Storage/State:**
    * eBPF Maps: Storing authorized process-ids and monotonic session tokens.

## 5. Alternatives Considered
* **Pure Software Validation**: Rejected due to vulnerability to browser-based SOP bypasses.
* **Network Namespacing (Docker)**: Effective but high overhead for local non-containerized developer workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HELI is a foundational component of the Zero-Trust local transport pillar.
* **Observability:** Kernel-level logs of dropped packets will be surfaced to the Local Security Violation Monitor.

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
