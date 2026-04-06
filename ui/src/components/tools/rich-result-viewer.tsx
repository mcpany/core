/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SmartResultRenderer } from "./smart-result-renderer";

interface RichResultViewerProps {
    result: any;
}

/**
 * Intent: Document RichResultViewer
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * RichResultViewer delegates to SmartResultRenderer for a unified rendering experience.
 *
 * @param props - The component props.
 * @param props.result - The raw result object from the tool execution.
 * @returns The rendered component.
 */
export function RichResultViewer({ result }: RichResultViewerProps) {
    return <SmartResultRenderer result={result} />;
}
