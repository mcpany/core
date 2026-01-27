/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { SlidersHorizontal, AlignJustify, AlignLeft } from "lucide-react";
import { useViewPreferences } from "@/contexts/view-preferences";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from "@/components/ui/dropdown-menu";

/**
 * A dropdown menu to configure view options like density.
 */
export function ViewOptions() {
  const { density, setDensity } = useViewPreferences();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" title="View Options">
          <SlidersHorizontal className="h-[1.2rem] w-[1.2rem]" />
          <span className="sr-only">View options</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>View Options</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">Density</DropdownMenuLabel>
        <DropdownMenuRadioGroup value={density} onValueChange={(v) => setDensity(v as "comfortable" | "compact")}>
            <DropdownMenuRadioItem value="comfortable" className="cursor-pointer">
                <AlignJustify className="mr-2 h-4 w-4 opacity-50" /> Comfortable
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="compact" className="cursor-pointer">
                <AlignLeft className="mr-2 h-4 w-4 opacity-50" /> Compact
            </DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
