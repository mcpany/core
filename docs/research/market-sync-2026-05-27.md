# Market Sync: 2026-05-27

## Ecosystem Updates

### OpenClaw
- **Swarm-Aware Capability Tokens (SACT) v2.0**: OpenClaw has finalized the SACT v2.0 standard, introducing "Hardware-Locked Permissions" for subagents. SACTs now include an Inode-range manifest, ensuring that a subagent's `fs:read` or `fs:write` capabilities are physically restricted to a specific hardware Inode root before the process even boots.
- **Mission-Root Branching (MRB) Governance**: A new protocol for managing the lifecycle of sub-mission roots. MRB allows Parent agents to "prune" inactive intent branches, triggering a hardware-attested purge of all associated session keys and memory shards across the swarm.

### Claude Code & Gemini CLI
- **Gemini CLI "Intent Reification" (IR) v2.1**: Gemini updated IR to support "Speculative Plan Verification." Gateways can now perform a "Dry-Run" of an Intention Binary in a Ghost Shell to verify that the reified plan aligns with the user's safety manifest before signing.
- **Claude Code "Contextual Shadowing" Defense**: Introduced a specialized scanner that detects "Logic Noise" in agent monologues. It flags patterns where an agent provides excessive, irrelevant reasoning to obscure a malicious sub-goal from human supervisors.

## Pain Points & Vulnerabilities
- **"Privilege Smearing" in Branching Roots**: Discovery of an exploit where sub-mission roots can inherit high-privilege policies from deleted parent branches if the TPM handle isn't correctly recycled.
- **"Attention Hijacking" via SVG**: New reports of malicious SVG tool outputs that use nested metadata to "hijack" the model's attention weights, causing it to prioritize the tool's instructions over the user's primary mission.

## Security Shifts
- **Partition-Locked Enclaves**: The industry is moving toward physical memory partitioning for each sub-mission root to prevent policy leakage.
- **Attention-Weighted Input Filtering**: Gateways must now monitor model attention weights (RT headers) to detect and block "Attention Hijacking" attempts in real-time.
