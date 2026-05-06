# MCP Any - Aesthetic Spec: Premium Stack Matrix

## 1. Goal
Transform the MCP Any experience by introducing a "Premium Stack Matrix" visualizer. This feature will act as a "Portainer for MCP" experience, bringing a sleek, card-based, ultra-premium interface to manage MCP Services (Stacks), Tools, and Resources.

## 2. Friction Point & Justification
Currently, MCP Any provides functional lists and basic graph topology views for connected services and tools. However, it lacks a dense, beautifully structured "App Grid" or "Stack Matrix" that users of systems like Portainer expect. A premium developer experience demands a centralized command center where users can instantly see the health, resource utilization (CPU/Memory proxy via QPS/Latency), and status of all their connected MCP servers in a highly polished, interactive grid.

## 3. Aesthetic Spec

### Colors & Typography
*   **Background:** Deep slate `#0f172a` ensuring a premium dark mode feel.
*   **Cards (Stacks):** Frosted glass panels with a slight border (`border-slate-800 bg-slate-900/50 backdrop-blur-xl`).
*   **Status Indicators:**
    *   **Healthy:** Neon emerald `#10b981` with a soft glow (`drop-shadow-[0_0_5px_rgba(16,185,129,0.5)]`).
    *   **Degraded:** Vibrant amber `#f59e0b`.
    *   **Offline/Error:** Sharp crimson `#ef4444`.
*   **Typography:** System sans-serif (Inter) with varying weights. Titles should be `font-semibold text-slate-100`, while metrics are `font-mono text-slate-400`.

### Interactions & Animations
*   **Hover States:** When hovering over a stack card, it should elevate smoothly (`translate-y-[-2px] shadow-lg shadow-cyan-500/10`) with a subtle border highlight (`border-slate-700`).
*   **Live Sparklines:** Each card should feature a minimalist SVG sparkline illustrating real-time QPS/Latency, updating fluidly without layout shifts.
*   **Click-to-Flip/Expand:** Clicking a card smoothly expands it into a detailed view or flips it to reveal the specific tools and resources bound to that stack.

### Core Components
*   **The Grid Layout:** A responsive CSS Grid (`grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4`) adjusting perfectly to screen sizes.
*   **Stack Card:** The primary unit of the matrix, containing the Service Icon, Name, Uptime, Live Sparkline, and a summary badge of active Tools/Resources.

## 4. Implementation Readiness
*   **Component Creation:** Create `ui/src/components/dashboard/premium-stack-matrix.tsx` to implement this visual grid.
*   **Seeding Update:** Enhance `server/pkg/app/seeds.go` to include a `premium-stack-matrix` mock service. Update e2e seeders to inject realistic, high-fidelity traffic, tool counts, and latency stats to populate the sparklines perfectly upon initial deployment.
