/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Button } from "@/components/ui/button";
import { X, Play, Pause, Star, StarOff, Loader2 } from "lucide-react";

interface BulkActionsToolbarProps {
  selectedCount: number;
  onAction: (action: 'enable' | 'disable' | 'pin' | 'unpin' | 'clear') => void;
  isProcessing: boolean;
}

export function BulkActionsToolbar({ selectedCount, onAction, isProcessing }: BulkActionsToolbarProps) {
  if (selectedCount === 0) return null;

  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-popover/95 backdrop-blur-md border shadow-lg rounded-full px-4 py-2 flex items-center gap-2 z-50 animate-in slide-in-from-bottom-5 fade-in duration-300">
      <div className="text-sm font-medium border-r pr-4 mr-2">
        {selectedCount} Selected
      </div>
      <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onAction('enable')}
            disabled={isProcessing}
            title="Enable Selected"
            className="h-8 px-2 rounded-full"
          >
              <Play className="h-4 w-4 mr-1 text-green-500" /> Enable
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onAction('disable')}
            disabled={isProcessing}
            title="Disable Selected"
             className="h-8 px-2 rounded-full"
          >
              <Pause className="h-4 w-4 mr-1 text-amber-500" /> Disable
          </Button>
          <div className="w-px h-4 bg-border mx-1" />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onAction('pin')}
            disabled={isProcessing}
            title="Pin Selected"
             className="h-8 px-2 rounded-full"
          >
              <Star className="h-4 w-4 mr-1 text-yellow-400" /> Pin
          </Button>
           <Button
            variant="ghost"
            size="sm"
            onClick={() => onAction('unpin')}
            disabled={isProcessing}
            title="Unpin Selected"
             className="h-8 px-2 rounded-full"
          >
              <StarOff className="h-4 w-4 mr-1" /> Unpin
          </Button>
      </div>
      <div className="border-l pl-2 ml-2">
         <Button
            variant="ghost"
            size="icon"
            onClick={() => onAction('clear')}
            disabled={isProcessing}
             className="h-8 w-8 rounded-full hover:bg-muted"
          >
              <X className="h-4 w-4" />
          </Button>
      </div>
      {isProcessing && <Loader2 className="h-4 w-4 animate-spin ml-2 text-primary" />}
    </div>
  );
}
