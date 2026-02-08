/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { Loader2 } from "lucide-react";

/**
 * StackEditPage component.
 * @returns The rendered component.
 */
export default function StackEditPage() {
  const params = useParams();
  const name = typeof params.name === 'string' ? decodeURIComponent(params.name) : "";

  const [stack, setStack] = useState<ServiceCollection | undefined>(undefined);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!name) return;

    async function load() {
        try {
            const data = await apiClient.getCollection(name);
            setStack(data);
        } catch (e) {
            console.error("Failed to load stack", e);
        } finally {
            setLoading(false);
        }
    }
    load();
  }, [name]);

  if (loading) {
      return (
          <div className="flex items-center justify-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (!stack) {
      return (
          <div className="flex items-center justify-center h-full text-muted-foreground">
              Stack not found.
          </div>
      );
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <StackEditor initialStack={stack} />
    </div>
  );
}
