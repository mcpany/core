/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { PromptWorkbench } from "@/components/prompts/prompt-workbench";

/**
 * Summary: PromptsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
export default function PromptsPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8">
      <PromptWorkbench />
    </div>
  );
}
