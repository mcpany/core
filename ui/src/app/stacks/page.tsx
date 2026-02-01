/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { StackList } from "@/components/stacks/stack-list";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" /> Create Stack
        </Button>
      </div>

      <StackList key={refreshTrigger} />

      <CreateStackDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        onComplete={() => setRefreshTrigger(prev => prev + 1)}
      />
    </div>
  );
}
