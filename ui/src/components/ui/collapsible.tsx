/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as CollapsiblePrimitive from "@radix-ui/react-collapsible"

/**
 * An interactive component which expands/collapses a panel.
 */
const Collapsible = CollapsiblePrimitive.Root

/**
 * The trigger that toggles the collapsible.
 */
const CollapsibleTrigger = CollapsiblePrimitive.CollapsibleTrigger

/**
 * The content that is shown/hidden by the collapsible.
 */
const CollapsibleContent = CollapsiblePrimitive.CollapsibleContent

export { Collapsible, CollapsibleTrigger, CollapsibleContent }
