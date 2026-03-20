import os

filepath = 'docs/features/design-hardware-locked-coordination-handshake.md'
with open(filepath, 'r') as f:
    lines = f.readlines()

new_content = [
    "# Design Doc: Hardware-Locked Coordination Handshake\n",
    "**Status:** Draft\n",
    "**Created:** 2026-06-14\n",
    "\n",
    "## 1. Context and Scope\n",
    "The emergence of Identity-Decay Attacks (IDA) and session-hijacking vulnerabilities in the A2A protocol (CVE-2026-48201) necessitates a transition from software-only handshakes to hardware-locked sessions. Existing mechanisms rely on long-lived tokens that can be mimicked or recycled.\n",
    "\n",
    "The Hardware-Locked Coordination Handshake (HLCH) mandates that every inter-agent task bidding and state fragment transfer be bound to a unique, TPM-signed session key. This ensures that even if an agent's behavioral trust is established, its coordination remain anchored to a non-repudiable hardware root.\n",
    "\n",
    "## 2. Goals & Non-Goals\n",
    "* **Goals:**\n",
    "    * Implement HLCH v1.0 standard for all inter-agent coordination.\n",
    "    * Mandate hardware-bound (TPM/Secure Enclave) signatures for coordination session keys.\n",
    "    * Neutralize nonce-recycling and stylometric mimicry (Identity-Decay).\n",
    "    * Provide collision-resistant attestation for reasoning fragments.\n",
    "* **Non-Goals:**\n",
    "    * Replacing primary user authentication (this is for inter-agent coordination).\n",
    "    * Managing persistent storage encryption (handled by HAPE/RAMS).\n",
    "\n",
    "## 3. Critical User Journey (CUJ)\n",
    "* **User Persona:** Multi-Agent Swarm Orchestrator\n",
    "* **Primary Goal:** Prevent a specialized subagent from performing an Identity-Decay attack via stylometric mimicry during a long-running session.\n",
    "* **The Happy Path (Tasks):**\n",
    "    1. Parent Agent initiates a task delegation request to Teammate A.\n",
    "    2. Teammate A generates a session-bound coordination key pair.\n",
    "    3. Teammate A requests a hardware-attestation signature for the session key from the local MRA Provider.\n",
    "    4. MRA Provider utilizes the TPM to sign the session key and mission-root fragment.\n",
    "    5. Teammate A completes the HLCH handshake with the parent via the T2T Bridge.\n",
    "    6. All subsequent state fragments from Teammate A must be signed by the hardware-locked session key.\n",
    "    7. Parent Agent (via SCI) verifies the hardware signature for every fragment, neutralizing mimicry.\n",
    "\n",
    "## 4. Design & Architecture\n",
    "* **System Flow:**\n",
    "    * Inter-agent request triggers session key generation.\n",
    "    * MRA Provider issues TPM-signed attestation for the key.\n",
    "    * HLCH Gateway validates the handshake before establishing the T2T pipe.\n",
    "    * SCI Interceptor enforces hardware-bound signatures on all transit metadata.\n",
    "* **APIs / Interfaces:**\n",
    "    * `hlch.InitiateHandshake(missionRoot, targetAgent) -> HandshakeToken`: Begins the hardware-locked handshake.\n",
    "    * `hlch.VerifySessionSignature(fragment, signature) -> bool`: Validates fragment integrity against the session key.\n",
    "* **Data Storage/State:**\n",
    "    * **Handshake Registry:** Ephemeral, memory-mapped storage for active hardware-locked session metadata.\n",
    "\n",
    "## 5. Alternatives Considered\n",
    "* **Software-Only Nonces:** Rejected due to vulnerability to CVE-2026-48201 (nonce-recycling).\n",
    "* **Per-Fragment TPM Signatures:** Rejected due to the high latency (100ms+) of repeated hardware calls; HLCH utilizes hardware-locked session keys for performance.\n",
    "\n",
    "## 6. Cross-Cutting Concerns\n",
    "* **Security (Zero Trust):** Mandatory hardware root of trust for all coordination. HLCH ensures that session keys are non-exportable from the TPM.\n",
    "* **Observability:** Integrated with the \"Coordination Handshake Debugger\" for real-time visualization of attestation events.\n",
    "\n",
    "## 7. Evolutionary Changelog\n",
    "* **2026-06-14:** Initial Document Creation. Supporting the transition to Hardware-Locked Coordination.\n"
]

with open(filepath, 'w') as f:
    f.writelines(new_content)
