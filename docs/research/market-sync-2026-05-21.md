# Market Sync: 2026-05-21

## Ecosystem Updates

### OpenClaw
- **Swarm-Aware Capability Tokens (SACT)**: OpenClaw has proposed SACT, a new standard for Task-Bound Resource Isolation (TBRI). SACTs are cryptographically bound to a specific mission-root and can be "delegated" to subagents with hardware-enforced restrictions on tool usage and compute.
- **Cognitive Isolation Quorums (CIQ)**: Expansion of SCQ that allows for "Peer-Review" of speculative reasoning zones before they are merged into the safe zone. CIQ uses a "Zero-Knowledge" approach where auditors can verify intent alignment without seeing the raw speculative data.

### Claude Code & Gemini CLI
- **Gemini CLI "Intent Entropy" (IE) Metrics**: Gemini now reports IE scores, which quantify the "surprise" or divergence of an agent's latest monologue from its initial instructions. High IE scores trigger an automatic "Intent Re-Alignment" (IRA) handshake in compliant gateways.
- **Claude Code "Inode-Locked Workspaces"**: A new filesystem protection that binds an agent's entire workspace to a hardware Inode-root. Any attempt to access files outside this root, even via complex symlink tunnels, results in an immediate security fault.

## Pain Points & Vulnerabilities
- **"Reasoning Smuggling" via EXIF**: New reports of malicious images where CoT instructions are embedded in EXIF data. Agents performing multimodal analysis ingest these instructions, bypassing textual CoT Integrity Shielding.
- **"Lease Racing" in TPM Slots**: In large swarms, subagents are "racing" to acquire the last available hardware-protected intent slots, leading to denial-of-service for mission-critical audit agents.

## Security Shifts
- **Full-Spectrum Intent Shielding**: Move toward analyzing ALL data fragments (multimodal and textual) for imperative reasoning signals.
- **Hardware Slot Fair-Share**: Gateways must implement "Priority Queues" for TPM slots to ensure supervisors always have a hardware anchor.
