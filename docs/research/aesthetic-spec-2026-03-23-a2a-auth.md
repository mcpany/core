# MCP Any - Aesthetic Spec: A2A Auth Status Dashboard

## 1. Goal

Introduce the "A2A Auth Status Dashboard," a premium, "Apple-level" interactive visualization for monitoring authenticated inter-agent handshakes and peer-to-peer security status across multi-agent swarms. This satisfies the P0 requirement from the 2026-03-23 Evolution.

## 2. Friction Point & Justification

As multi-agent swarms become more complex, tracking the cryptographic sovereignty and authentication state between communicating subagents is opaque. Developers need a centralized, visually stunning dashboard to instantly verify which agents are securely connected, which handshakes are pending, and where security boundaries might be compromised. A standard list view is insufficient; a dynamic, topology-aware interface is required to provide immediate visual confirmation of the mesh's security posture. If it doesn't look like a high-end mission control center, it's a bug.

## 3. Customer User Journey (CUJ)

1.  **Dashboard Entry:** The user navigates to the "A2A Auth" tab within the MCP Any dashboard.
2.  **Global Mesh View:** They are presented with a real-time, interactive graph or high-fidelity grid showing all active subagents and their current handshake states (Authenticated, Pending, Failed).
3.  **Deep Dive:** Hovering over a specific agent-to-agent connection reveals the cryptographic details of the handshake (e.g., TPM-signed tokens, lease duration).
4.  **Anomaly Detection:** Any failed or unauthorized handshake attempts are immediately highlighted with pulsing destructive indicators, drawing the user's eye to potential security events.

## 4. Aesthetic Spec & Vibe

### Colors & Typography
*   **Background:** Deep Space (Slate-950) with subtle radial gradients to create depth (`bg-slate-950 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-slate-900 to-slate-950`).
*   **Authenticated State:** A vibrant, stable Cyan (`text-cyan-400`) accompanied by a subtle neon glow (`drop-shadow-[0_0_8px_rgba(34,211,238,0.5)]`).
*   **Pending/Negotiating State:** A pulsing Amber (`text-amber-400` with `animate-pulse`).
*   **Failed/Unauthorized State:** A stark, sharp Crimson (`text-rose-500`) with a harsh geometric border.
*   **Typography:** Primary UI elements use Inter. Cryptographic hashes and token previews strictly use a monospace font (JetBrains Mono preferred) at a reduced size (`text-[10px] text-muted-foreground/80`).

### Interactions & Animations
*   **Connection Lines (if graph view):** Animated dashed lines showing the direction of the authentication request, with the "marching ants" effect (`stroke-dasharray` animation).
*   **Card Hover:** Smooth elevation and a slight border glow transition on agent cards (`transition-all duration-300 hover:shadow-[0_0_15px_rgba(34,211,238,0.2)] hover:border-cyan-500/30`).
*   **Handshake Success:** A brief ripple effect or flash of cyan when a pending handshake successfully resolves.

### Core Components
*   **The Identity Node:** A visual representation of an agent, featuring a circular avatar/icon ringed by its current authentication status color.
*   **The Handshake Inspector:** A slide-out panel or rich tooltip that displays the full JSON payload of the attestation token when a connection is clicked.
*   **The Sovereignty Ring:** A visual indicator (like a concentric circle) around groups of agents that share the same hardware-locked environment sovereignty.

## 5. Implementation Readiness

*   **Design:** This spec is final and requires zero human opinion.
*   **Seeding Update:** `server/pkg/app/seeds_collections.go` (and related seeders) must be updated to inject mock "Gold Standard" A2A authentication data, including successful handshakes and simulated failures, so the dashboard populates beautifully upon first load.
