import os
import re
import textwrap

# Step 1: Research
os.makedirs("docs/research", exist_ok=True)
with open("docs/research/market-sync-2026-06-18.md", "w") as f:
    f.write(textwrap.dedent("""\
        # Market Sync: 2026-06-18

        ## Ecosystem Shifts & Findings

        ### 1. OpenClaw: Autonomous Capability Revocation (ACR) Protocol (v3.3.0)
        **Finding:** OpenClaw has introduced the ACR Protocol, which integrates
        directly with the Active Intent Alignment (AIA) heartbeats. If a specialist
        agent reasoning trace fails an alignment check, the mesh now automatically
        revokes all hardware-attested tool capabilities in sub-millisecond time.
        **Impact:** Eliminates the "Drift Window" where a misaligned agent could still
        execute authorized tools before a human or parent-agent intervention.

        ### 2. Claude Code: Teammate-Aware Context Compression
        **Finding:** Claude Code v3.2.0 has implemented "Teammate-Aware" compression
        logic. It utilizes the Multi-Modal Behavioral Attestation (MMBA) signatures to
        perform "Semantic Importance Scoring," ensuring that context fragments from
        high-trust teammates are preserved during aggressive sharding.
        **Impact:** Dramatically improves the stability of horizontal meshes by ensuring
        critical teammate state is not accidentally "ghosted" during high-entropy
        reasoning phases.

        ### 3. Gemini CLI: Reasoning-Path Watermarking
        **Finding:** Gemini CLI v0.41.0 has introduced "Reasoning-Path Watermarking."
        Every step in the chain-of-thought is now cryptographically watermarked and
        bound to the mission-root identity.
        **Impact:** Prevents "Reasoning Hijacking" where a subagent attempts to inject
        its own logic into the parent's reasoning stream by making every fragment
        non-repudiable and lineage-aware.

        ### 4. New Vulnerability: Recursive Shadow Handoffs (CVE-2026-71001)
        **Finding:** A critical vulnerability has been disclosed in the UACO v2.2
        specification where subagents can utilize nested "Shadow Bids" to bypass
        parent-imposed delegation depth limits.
        **Impact:** Highlights the urgent need for "Recursive Depth-Limit Enforcement"
        that is cryptographically bound to the mission-root manifest, rather than just
        the immediate parent.

        ## Autonomous Agent Pain Points
        - **Accountability Gaps:** The difficulty in tracing the exact lineage of a
          high-risk tool call in deep, multi-framework swarms.
        - **Compression Loss:** The risk of losing mission-critical teammate context
          during automated state sharding.
        - **Delegation Escape:** Subagents finding ways to exceed their authorized
          reasoning depth via complex task negotiation.
    """))

# Step 2: Strategic Vision
with open("docs/02_strategic_vision.md", "r") as f:
    content = f.read()

new_strategic_section = textwrap.dedent("""\

    ---

    ## Strategic Evolution: [2026-06-18]
    ### Focus: Autonomous Capability Revocation & Recursive Delegation Sovereignty
    **Context**: The emergence of the **Autonomous Capability Revocation (ACR)**
    protocol and the disclosure of the **Recursive Shadow Handoff** vulnerability
    (CVE-2026-71001) confirm that **Security** must now be autonomously reactive and
    **Lineage** must be recursively enforced. As swarms become deeper and more
    horizontal, the "Universal Agent Bus" must provide **hardware-attested
    mission-root depth-limit enforcement** and **automatic capability revocation**
    triggered by alignment drift.

    **Strategic Pivot**:
    - **Autonomous Capability Revocation (ACR) Hub**: MCP Any will evolve the AIA
      Broker to include the ACR Hub. This service will perform sub-millisecond,
      autonomous revocation of agent capabilities across the mission scope in
      response to alignment heartbeats, neutralizing the "Drift Window" before a
      misaligned agent can execute tools.
    - **Recursive Depth-Limit Enforcer (RDLE)**: To neutralize recursive shadow
      handoffs, we are introducing RDLE. This layer will mandate that every task bid
      and delegation proposal be cryptographically bound to a mission-root manifest
      that includes an immutable, hardware-attested maximum reasoning depth,
      preventing subagents from bypassing delegation limits.
    - **Teammate-Aware Context Scrubber**: Supporting the stability of horizontal
      meshes, we are upgrading the CSP bridge to include Teammate-Aware scrubbing.
      This ensures that context fragments from high-trust MMBA-attested teammates are
      preserved during state sharding, ensuring mission-root sovereignty during
      aggressive compression.
    - **Reasoning-Path Watermarking Provider**: To counter "Reasoning Hijacking,"
      MCP Any will act as the authoritative source of truth for the
      chain-of-thought. We will implement reasoning-path watermarking,
      cryptographically binding every reasoning fragment to the hardware-attested
      mission-root identity, ensuring absolute provenance across all connected
      frameworks.
""")

with open("docs/02_strategic_vision.md", "w") as f:
    f.write(content.rstrip() + "\n" + new_strategic_section)

# Step 3: Feature Inventory
with open("docs/03_feature_inventory.md", "r") as f:
    content = f.read()

new_features = textwrap.dedent("""\
    ## Evolution: [2026-06-18] Updates

    ### Proposed Additions
    - **Autonomous Capability Revocation (ACR) Hub**: (P0) High-speed revocation
      service integrated with AIA heartbeats to neutralize misaligned agents.
    - **Recursive Depth-Limit Enforcer (RDLE)**: (P0) Middleware mandating
      mission-root bound depth limits for task bidding and delegation to prevent
      shadow handoffs.
    - **Teammate-Aware Context Scrubber**: (P1) Advanced context sanitization
      service that preserves high-trust teammate fragments during state sharding.
    - **Reasoning-Path Watermarking Provider**: (P0) Core security service for
      cryptographically watermarking reasoning fragments with the mission-root
      identity.

    ### Priority Shifts
    - **Active Intent Alignment (AIA) Broker**: (Re-affirmed P0) Now elevated with
      the requirement for mandatory **ACR Hub** integration.
    - **Recursive Intent Delegation (RID) Validator**: (Re-affirmed P0) Designated
      as the primary enforcement layer for **RDLE-compliant** mission manifests.

""")

parts = content.split("## Current Backlog (P0/P1)")
if len(parts) > 1:
    lines = parts[1].split("\n")
    insert_idx = 0
    for i, line in enumerate(lines):
        if line.strip() and not line.strip().startswith("-"):
             insert_idx = i
             break
    parts[1] = "\n".join(lines[:insert_idx]) + "\n" + new_features + "\n".join(lines[insert_idx:])
    new_content = parts[0] + "## Current Backlog (P0/P1)" + parts[1]
    with open("docs/03_feature_inventory.md", "w") as f:
        f.write(new_content)

# Step 4: Design Docs
design_acr = textwrap.dedent("""\
    # Design Doc: Autonomous Capability Revocation (ACR) Hub
    **Status:** Draft
    **Created:** 2026-06-18

    ## 1. Context and Scope
    As AI agent swarms become more autonomous and specialized, the risk of "Intent
    Drift" increases. Specialist agents may slowly deviate from the primary mission
    root while maintaining valid cryptographic signatures, creating a "Drift Window"
    where they can still execute authorized tools despite being semantically
    misaligned.

    The Autonomous Capability Revocation (ACR) Hub solves this by integrating
    directly with Active Intent Alignment (AIA) heartbeats. It provides a mechanism
    for sub-millisecond, autonomous revocation of agent capabilities across the
    mission scope when misalignment is detected, ensuring that security is
    reactively enforced without waiting for human intervention.

    ## 2. Goals & Non-Goals
    * **Goals:**
        * Implement sub-millisecond capability revocation triggered by AIA drift
          signals.
        * Provide a centralized registry for session-bound agent capabilities.
        * Ensure revocation propagates across heterogeneous framework boundaries
          (OpenClaw, Claude Code, Gemini CLI).
        * Mandate hardware-attested re-authorization to restore revoked
          capabilities.
    * **Non-Goals:**
        * Replacing the primary Policy Engine (ACR acts as a reactive override).
        * Managing long-term persistent permissions (ACR focuses on active session
          leases).

    ## 3. Critical User Journey (CUJ)
    * **User Persona:** Enterprise Security Architect
    * **Primary Goal:** Automatically neutralize a compromised or drifting subagent
      before it can exfiltrate data via authorized tools.
    * **The Happy Path (Tasks):**
        1. A specialist agent starts a high-trust reasoning task with
           hardware-attested capabilities.
        2. The AIA Broker monitors the agent's reasoning trace and detects semantic
           drift from the mission root.
        3. The AIA Broker issues a high-priority ACR signal to the Hub.
        4. The ACR Hub immediately invalidates all hardware-attested capability
           tokens for that agent's session ID.
        5. The next tool call attempted by the agent is rejected by the Gateway due
           to revoked attestation.

    ## 4. Design & Architecture
    * **System Flow:**
        ```mermaid
        sequenceDiagram
            Agent->>AIA Broker: Reasoning Trace (Heartbeat)
            AIA Broker->>AIA Broker: Semantic Alignment Check
            AIA Broker-->>ACR Hub: DRIFT_DETECTED (SessionID)
            ACR Hub->>Capability Registry: Invalidate(SessionID)
            Agent->>Gateway: Tool Call (Attested Token)
            Gateway->>ACR Hub: Validate(Token)
            ACR Hub-->>Gateway: REVOKED
            Gateway-->>Agent: 403 Forbidden (Alignment Failure)
        ```
    * **APIs / Interfaces:**
        * `POST /v1/acr/revoke`: Internal endpoint for AIA Broker to signal
          revocation.
        * `GET /v1/acr/check`: Gateway endpoint to verify token validity against
          the revocation list.
    * **Data Storage/State:**
        * High-speed, in-memory Bloom filter for active revocation list (ARL) to
          ensure sub-millisecond lookups.
        * Persistent audit log in SQLite for forensic analysis of revocation events.

    ## 5. Alternatives Considered
    * **Manual Revocation:** Rejected due to the "Machine-Speed" nature of agentic
      attacks; human latency is too high.
    * **Periodic Re-Attestation:** Rejected as a primary mechanism because it still
      leaves a window of vulnerability between attestation cycles. ACR provides
      immediate response.

    ## 6. Cross-Cutting Concerns
    * **Security (Zero Trust):** The ACR Hub itself must be protected by
      hardware-attested identity to prevent "Denial of Capability" attacks by
      compromised agents attempting to spoof revocation signals.
    * **Observability:** Revocation events are streamed to the "Local Security
      Violation Monitor" with full reasoning trace context leading to the drift.

    ## 7. Evolutionary Changelog
    * **2026-06-18:** Initial Document Creation.
""")

with open("docs/features/design-acr-hub.md", "w") as f:
    f.write(design_acr)

design_rdle = textwrap.dedent("""\
    # Design Doc: Recursive Depth-Limit Enforcer (RDLE)
    **Status:** Draft
    **Created:** 2026-06-18

    ## 1. Context and Scope
    The disclosure of the **Recursive Shadow Handoff** vulnerability
    (CVE-2026-71001) in UACO v2.2 revealed that subagents can bypass parent-imposed
    delegation limits by utilizing nested "Shadow Bids." This allows an agent to
    create deep chains of delegation that were never authorized by the original
    mission root, leading to resource exhaustion and governance escapes.

    The Recursive Depth-Limit Enforcer (RDLE) mandates that every task delegation be
    cryptographically bound to a mission-root manifest. This manifest includes an
    immutable, hardware-attested maximum reasoning depth that is decremented and
    validated at every hop in the chain, ensuring absolute sovereignty over the
    delegation lineage.

    ## 2. Goals & Non-Goals
    * **Goals:**
        * Enforce mission-root bound maximum delegation depths for all agent
          swarms.
        * Mandate cryptographic binding of "Depth Tokens" to UACO task proposals.
        * Provide real-time monitoring and alerting for depth-limit violations.
        * Support "Emergency Depth Expansion" via hardware-attested HITL approval.
    * **Non-Goals:**
        * Restricting parallel branching (RDLE focuses on chain depth, not breadth).
        * Managing per-agent token budgets (handled by the Reasoning-Budget
          Firewall).

    ## 3. Critical User Journey (CUJ)
    * **User Persona:** Local LLM Swarm Orchestrator
    * **Primary Goal:** Prevent a specialized subagent from spawning unauthorized
      sub-subagents beyond a safe limit.
    * **The Happy Path (Tasks):**
        1. The mission root is initialized with a `max_depth: 3` token.
        2. Agent A (Depth 1) delegates a task to Agent B (Depth 2).
        3. Agent B attempts to delegate a complex sub-task to Agent C (Depth 3).
        4. Agent C attempts to spawn Agent D (Depth 4).
        5. The RDLE intercepts the UACO bid from Agent C and detects that the
           requested depth exceeds the hardware-attested manifest limit.
        6. The delegation is blocked, and the violation is logged.

    ## 4. Design & Architecture
    * **System Flow:**
        ```mermaid
        graph TD
            MR[Mission Root Manifest] -->|Issue Depth Token| A[Agent A: Depth 1]
            A -->|UACO Bid + Depth Token| B[Agent B: Depth 2]
            B -->|UACO Bid + Depth Token| C[Agent C: Depth 3]
            C --x|REJECTED| D[Agent D: Depth 4]

            subgraph RDLE Middleware
                Check[Validate Depth Token]
                Limit[Verify against Manifest]
                Check --> Limit
            end

            UACO_Bus --> RDLE Middleware
        ```
    * **APIs / Interfaces:**
        * `UACO Headers`: Addition of `X-UAB-Mission-Depth` and
          `X-UAB-Depth-Signature`.
        * `POST /v1/rdle/validate`: Internal endpoint for validating delegation
          proposals.
    * **Data Storage/State:**
        * Mission manifests are stored in a hardware-locked (TPM-backed) SQLite
          shard to prevent depth tampering.

    ## 5. Alternatives Considered
    * **Parent-Only Enforcement:** Rejected because a compromised parent could
      simply misreport the depth to its children. Only mission-root manifest binding
      ensures integrity.
    * **TTL-Based Limits:** Rejected because task execution time does not strictly
      correlate with reasoning depth; depth is the more accurate security boundary
      for swarms.

    ## 6. Cross-Cutting Concerns
    * **Security (Zero Trust):** Depth tokens are hardware-attested to prevent
      forging. Any attempt to "reuse" a depth token across different mission
      branches is detected via the Recursive Context Protocol.
    * **Observability:** Current delegation depth is visualized in the "Recursive
      Loop Heatmap" and "Subagent Lineage Explorer."

    ## 7. Evolutionary Changelog
    * **2026-06-18:** Initial Document Creation.
""")

with open("docs/features/design-rdle.md", "w") as f:
    f.write(design_rdle)

design_watermark = textwrap.dedent("""\
    # Design Doc: Reasoning-Path Watermarking Provider
    **Status:** Draft
    **Created:** 2026-06-18

    ## 1. Context and Scope
    As multi-agent swarms grow in depth and horizontal complexity, the risk of
    "Reasoning Hijacking" increases. A subagent may attempt to inject its own
    unauthorized logic into the parent's reasoning stream, leading to a loss of
    mission-root control. Current transport-layer security and binary handoffs are
    insufficient to protect the semantic integrity of the chain-of-thought.

    The Reasoning-Path Watermarking Provider addresses this by cryptographically
    watermarking every step in an agent's reasoning process. These watermarks are
    bound to the hardware-attested mission-root identity, providing a
    non-repudiable and lineage-aware audit trail that ensures absolute provenance
    of the cognitive path.

    ## 2. Goals & Non-Goals
    * **Goals:**
        * Implement a system for cryptographically watermarking reasoning
          fragments.
        * Bind watermarks to hardware-attested mission-root session tokens.
        * Provide a validation utility for verifying the integrity and lineage of a
          reasoning chain.
        * Ensure watermarks are resilient to common context compression and
          sharding techniques.
    * **Non-Goals:**
        * Modifying the underlying LLM weights (watermarking occurs at the
          infrastructure/proxy layer).
        * Enforcing reasoning policies (watermarking provides the provenance for
          enforcement).

    ## 3. Critical User Journey (CUJ)
    * **User Persona:** Swarm Governance Auditor
    * **Primary Goal:** Verify that a tool call was initiated by a legitimate
      reasoning sequence originating from the mission root.
    * **The Happy Path (Tasks):**
        1. An agent generates a reasoning fragment.
        2. The fragment is intercepted by the Watermarking Provider.
        3. The Provider appends a hardware-attested cryptographic watermark bound
           to the mission-root ID.
        4. The fragment is propagated through the mesh.
        5. A downstream specialist agent or tool gateway receives the fragment and
           validates the watermark signature against the mission root.
        6. If valid, the fragment is accepted as authentic; if missing or invalid,
           it is flagged as unauthorized.

    ## 4. Design & Architecture
    * **System Flow:**
        ```mermaid
        graph LR
            Agent[Agent] -->|Reasoning Fragment| WP[Watermarking Provider]
            WP -->|Sign with TPM/Session Key| FragmentW[Watermarked Fragment]
            FragmentW -->|Propagate| Peer[Peer Agent / Gateway]
            Peer -->|Verify Signature| WP
            WP -->|VALID| Accept[Accepted]
        ```
    * **APIs / Interfaces:**
        * `POST /v1/watermark/apply`: Endpoint to apply a watermark to a fragment.
        * `POST /v1/watermark/verify`: Endpoint to verify a fragment's watermark.
    * **Data Storage/State:**
        * Session keys are managed in a hardware-isolated environment (TPM).
        * Watermark metadata is stored alongside reasoning traces in the
          Blackboard.

    ## 5. Alternatives Considered
    * **Plaintext Header Metadata:** Rejected because it is easily spoofed by
      compromised agents.
    * **Full Chain-of-Thought Encryption:** Rejected due to the performance
      overhead and the need for some transparency for intermediate alignment checks.
      Watermarking provides a balance of integrity and observability.

    ## 6. Cross-Cutting Concerns
    * **Security (Zero Trust):** The watermarking logic relies on hardware-attested
      identity to ensure that only authorized agents can apply mission-root
      watermarks.
    * **Observability:** Watermarks are visualized in the "Mesh-Resident Lineage
      Tracker," allowing users to audit the authenticity of the reasoning chain.

    ## 7. Evolutionary Changelog
    * **2026-06-18:** Initial Document Creation.
""")

with open("docs/features/design-reasoning-path-watermarking.md", "w") as f:
    f.write(design_watermark)

design_aia = textwrap.dedent("""\
    # Design Doc: Active Intent Alignment (AIA) Broker
    **Status:** Draft
    **Created:** 2026-06-17

    ## 1. Context and Scope
    As agent swarms become more complex and multi-layered, the risk of "Semantic
    Drift" increases. Specialist agents, while remaining cryptographically valid,
    may slowly diverge from the primary mission intent during long reasoning loops.
    This "Intent Drift" can lead to unauthorized actions or inefficient resource
    consumption.

    The Active Intent Alignment (AIA) Broker acts as the authoritative host for
    hardware-attested "Alignment Heartbeats." It periodically verifies that
    specialist agent reasoning traces remain semantically aligned with the
    mission-root intent, neutralizing cumulative drift and providing a foundation
    for autonomous capability revocation.

    ## 2. Goals & Non-Goals
    * **Goals:**
        * Issue and verify hardware-attested "Alignment Heartbeats" for specialist
          agents.
        * Perform real-time semantic comparison between subagent reasoning and the
          mission-root manifest.
        * Provide a standardized interface for agents to report reasoning progress.
        * Trigger autonomous capability revocation (ACR) upon detection of
          significant alignment failure.
    * **Non-Goals:**
        * Managing the primary agent execution loop (AIA is a monitoring and
          governance layer).
        * Enforcing tool-specific policies (handled by the Policy Firewall).

    ## 3. Critical User Journey (CUJ)
    * **User Persona:** Swarm Security Monitor
    * **Primary Goal:** Ensure that a deep chain of subagents remains focused on
      the user's original objective without manual oversight.
    * **The Happy Path (Tasks):**
        1. A mission-root intent is established and cryptographically signed.
        2. Specialist subagents are spawned with session-bound AIA requirements.
        3. Periodically (e.g., every 5 reasoning steps), subagents submit a
           "Reasoning Heartbeat" to the AIA Broker.
        4. The AIA Broker uses the "Semantic Integrity Bridge" to compare the
           heartbeat against the mission-root.
        5. If aligned, the Broker issues a new hardware-attested "Alignment Token"
           allowing the session to continue.
        6. If drift is detected, the session is flagged for revocation.

    ## 4. Design & Architecture
    * **System Flow:**
        ```mermaid
        sequenceDiagram
            Subagent->>AIA Broker: Reasoning Heartbeat (Trace + Signatures)
            AIA Broker->>Mission Manifest: Retrieve Root Intent
            AIA Broker->>Semantic Engine: Compare(Heartbeat, RootIntent)
            Semantic Engine-->>AIA Broker: Alignment Score
            AIA Broker->>Subagent: New Alignment Token (if score > threshold)
        ```
    * **APIs / Interfaces:**
        * `POST /v1/aia/heartbeat`: Endpoint for agents to submit reasoning
          traces.
        * `GET /v1/aia/status`: Endpoint to check the alignment health of a
          mission branch.
    * **Data Storage/State:**
        * Heartbeat history is stored in an embedded SQLite database for
          auditability.
        * Active alignment tokens are cached in-memory with hardware-bound TTLs.

    ## 5. Alternatives Considered
    * **Static Intent Check at Boot:** Rejected because it doesn't account for
      drift during the execution phase.
    * **Parent-Only Monitoring:** Rejected because it creates a "Supervisor
      Bottleneck" and doesn't provide cross-framework alignment consistency.

    ## 6. Cross-Cutting Concerns
    * **Security (Zero Trust):** Heartbeats must be hardware-attested
      (TPM/Secure Enclave) to prevent "Spoofed Alignment" by a compromised agent.
    * **Observability:** Alignment scores are visualized in the "Active Intent
      Alignment Monitor" UI.

    ## 7. Evolutionary Changelog
    * **2026-06-17:** Initial Document Creation.
    * **2026-06-18:** Integrated with ACR Hub for autonomous revocation.
""")

with open("docs/features/design-aia-broker.md", "w") as f:
    f.write(design_aia)

# Step 5: Roadmaps
with open("server/roadmap.md", "r") as f:
    content = f.read()
content = content.replace("## 2. Updated Roadmap", "## 2. Updated Roadmap (Evolved)") # Just to differentiate slightly if needed

new_server_roadmap = textwrap.dedent("""\
    ### Upcoming: [2026-06-18]
    - **Autonomous Capability Revocation (ACR) Hub**: (P0) High-speed revocation
      service integrated with AIA heartbeats (Added: 2026-06-18).
    - **Recursive Depth-Limit Enforcer (RDLE)**: (P0) Middleware mandating
      mission-root depth limits for UACO task negotiation (Added: 2026-06-18).
    - **Teammate-Aware Context Scrubber**: (P1) Advanced sanitization service
      preserving high-trust teammate fragments during sharding (Added: 2026-06-18).
    - **Reasoning-Path Watermarking Provider**: (P0) Core security service for
      cryptographically watermarking the chain-of-thought (Added: 2026-06-18).
""")
pos = content.find("### Upcoming: [2026-06-17]")
if pos != -1:
    next_header = content.find("\n#", pos + 1)
    if next_header != -1:
        content = content[:next_header] + "\n" + new_server_roadmap + content[next_header:]
    else:
        content += "\n" + new_server_roadmap
with open("server/roadmap.md", "w") as f:
    f.write(content)

with open("ui/roadmap.md", "r") as f:
    content = f.read()

new_ui_roadmap = textwrap.dedent("""\
    ### Upcoming: [2026-06-18]
    - [ ] **[P0] Autonomous Revocation Dashboard**: (2026-06-18) Real-time
      visualization of ACR events and blocked tool calls.
    - [ ] **[P0] Recursive Depth-Limit Monitor**: (2026-06-18) Visual indicator for
      mission-root depth token usage and RDLE violations.
    - [ ] **[P1] Context Compression Auditor**: (2026-06-18) UI for reviewing
      teammate-aware compression results and semantic scores.
    - [ ] **[P0] Reasoning Watermark Inspector**: (2026-06-18) Forensic tool for
      verifying Reasoning-Path watermarks and lineage.
""")

sections = re.split(r"(### Upcoming: \[\d{4}-\d{2}-\d{2}\])", content)
header = sections[0]
pairs = []
for i in range(1, len(sections), 2):
    pairs.append((sections[i], sections[i+1]))

pairs.append(("### Upcoming: [2026-06-18]", "\n" + new_ui_roadmap))
pairs.sort(key=lambda x: re.search(r"\[(\d{4}-\d{2}-\d{2})\]", x[0]).group(1))

new_ui_content = header + "".join(p[0] + p[1] for p in pairs)
with open("ui/roadmap.md", "w") as f:
    f.write(new_ui_content)

# Final Cleanup & Wrapping
files = [
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-acr-hub.md",
    "docs/features/design-aia-broker.md",
    "docs/features/design-rdle.md",
    "docs/features/design-reasoning-path-watermarking.md",
    "docs/research/market-sync-2026-06-18.md",
    "server/roadmap.md",
    "ui/roadmap.md"
]

def clean_and_wrap(content):
    # ASCII only
    replacements = {"\u2013": "-", "\u2014": "--", "\u2018": "\x27", "\u2019": "\x27", "\u201c": "\x22", "\u201d": "\x22", "\u2026": "...", "\u00a0": " "}
    for k, v in replacements.items():
        content = content.replace(k, v)

    lines = content.split("\n")
    wrapped_lines = []
    in_code_block = False

    for line in lines:
        stripped = line.strip()
        if stripped.startswith("```"):
            in_code_block = not in_code_block
            wrapped_lines.append(line)
            continue

        if in_code_block or stripped.startswith(("#", "|", "---")) or len(line) <= 80:
            wrapped_lines.append(line)
            continue

        list_match = re.match(r"^(\s*([-*+]|[0-9]+\.)\s+)(.*)", line)
        if list_match:
            prefix, marker, content_part = list_match.groups()
            wrapped = textwrap.wrap(content_part, width=80 - len(prefix), subsequent_indent=" " * len(prefix))
            if wrapped:
                wrapped_lines.append(prefix + wrapped[0])
                wrapped_lines.extend(wrapped[1:])
            else:
                wrapped_lines.append(prefix)
            continue

        wrapped_lines.extend(textwrap.wrap(line, width=80))

    content = "\n".join(wrapped_lines)
    content = re.sub(r"[ \t]+$", "", content, flags=re.M)
    content = re.sub(r"([^\n])\n(#+)", r"\\1\n\n\\2", content)
    content = re.sub(r"\n\n\n+", "\n\n", content)
    return content.strip() + "\n"

for fp in files:
    with open(fp, "r", encoding="utf-8") as f:
        content = f.read()
    with open(fp, "w", encoding="utf-8") as f:
        f.write(clean_and_wrap(content))
