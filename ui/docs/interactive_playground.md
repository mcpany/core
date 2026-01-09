# Interactive Playground Feature

## Overview
The **Interactive Playground** transforms the tool testing experience into a polished, chat-like interface. It allows developers to execute registered tools against the backend (and simulated Next.js backend tools) in real-time, verifying inputs and outputs.

## Key Features
- **Real-time Tool Execution:** Directly execute tools via `POST /api/v1/execute`.
- **Chat Interface:** "Apple Intelligence" style chat bubbles with glassmorphism effects.
- **JSON Support:** Structured input and output display with syntax highlighting (via formatted `<pre>` blocks).
- **Tool Discovery:** A "Available Tools" sheet to quickly browse registered tools and their schemas.
- **Error Handling:** Clear error messages for invalid JSON or execution failures.
- **Built-in Tools:** Includes `calculator`, `echo`, `system_info`, and `weather` (mock) for immediate testing.

## Technical Details
- **Frontend:** Next.js Client Component (`PlaygroundClient`) using Tailwind CSS and Lucide icons.
- **Backend:** Next.js App Router API Routes (`/api/v1/execute`, `/api/v1/tools`) handling tool execution logic server-side.
- **Testing:** Unit tests for tool logic and Playwright E2E tests for the full chat flow.

## Screenshot
![Interactive Playground](interactive_playground.png)
