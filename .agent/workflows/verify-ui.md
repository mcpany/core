---
description: UI Hardening, Screenshot Automation & Verification
---

Role: You are an expert Full Stack Engineer and QA Specialist working on the "MCP Any" Next.js application.

Context: The CI/CD pipeline relies on accurate, error-free screenshots for documentation. Recent successful runs have revealed hydration errors, unhandled API states, and dark mode visibility issues that need permanent fixes.

Objective
Your goal is to harden the UI components against API failures, ensuring no "Application Error" or "Failed to load" states appear in screenshots. You must also automate the screenshot generation process using the existing Playwright suite and verify the results.

Instructions

1. Code Safety & Error Handling
   Audit the following key components and ensure they handle null, undefined, or non-array API responses gracefully (e.g., using optional chaining ?. and fallback || []):

Upstream Services:
ui/src/app/upstream-services/page.tsx
Tools Library:
ui/src/app/tools/page.tsx
Prompt Workbench:
ui/src/components/prompts/prompt-workbench.tsx
Resource Explorer:
ui/src/components/resources/resource-explorer.tsx
Global Search:
ui/src/components/global-search.tsx
Requirement: If the API returns a 401 or malformed data, the UI must render an empty state or a user-friendly error toast, NOT a Next.js crash screen.

2. Visual Improvements
   Dark Mode: Verify that all status indicators (e.g., "Offline", "Unknown") are visible against the dark background. specific attention to
   health-history-chart.tsx
   .
   Loading States: Ensure Dashboard widgets (Traffic, Failure Rate) have sufficient time to render before screenshots are captured. Increase page.waitForTimeout if necessary in tests.
3. Screenshot Automation
   You are provided with a dedicated Make target for regenerating screenshots.

Command: make update-screenshots (run this in core/ui).
Configuration:
ui/playwright.screenshots.config.ts
Test File:
ui/tests/generate_docs_screenshots.spec.ts
Action: Modify the test file to remove unnecessary mocks if "real" data (like System Logs resources) is available in the test environment. Prefer real data over mocks for authenticity.

4. Verification Workflow
   A. Automated Verification
   After applying fixes, run the screenshot generation suite:

cd core/ui
make update-screenshots
Check the output for any console warnings or test failures.

B. Manual UI Verification (Browser Subagent)
Use your browser capabilities to verify the fixes live:

Launch: Use the test container to start both frontend server and backend server.
Navigate: Go to http://localhost:3000/dashboard and http://localhost:3000/resources.
Inspect:
Check for console errors in the DevTools.
Verify that the "Offline" status bars are gray (visible), not black.
Confirm that the Resource Preview modal opens with content.
Resize the window to ensure responsiveness.
C. Artifact Inspection
Verify the timestamps and file sizes of the generated PNGs in
ui/docs/screenshots/
to confirm they were actually updated.

ls -l ui/docs/screenshots/dashboard_overview.png
ls -l ui/docs/screenshots/resource_preview_modal.png
Deliverables
Code Changes: Hardened React components.
Updated Screenshots: Fresh PNG files in
ui/docs/screenshots
.
Walkthrough: A summary of what was fixed and verified.
