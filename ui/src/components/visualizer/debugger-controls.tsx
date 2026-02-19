/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Button } from "@/components/ui/button";
import { Play, Pause, SkipForward, Square } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface DebuggerControlsProps {
  isPlaying: boolean;
  onPlayPause: () => void;
  onStep: () => void;
  onStop: () => void;
}

/**
 * DebuggerControls provides play/pause and step controls for the debugger.
 * @param props - The component props.
 * @param props.isPlaying - Whether the debugger is currently playing.
 * @param props.onPlayPause - Callback to toggle play/pause.
 * @param props.onStep - Callback to step forward.
 * @param props.onStop - Callback to stop execution.
 * @returns The DebuggerControls component.
 */
export function DebuggerControls({ isPlaying, onPlayPause, onStep, onStop }: DebuggerControlsProps) {
  return (
    <div className="flex items-center gap-1 bg-background border rounded-md p-1 shadow-sm">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onPlayPause}>
              {isPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {isPlaying ? "Pause" : "Resume"}
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onStep} disabled={isPlaying}>
              <SkipForward className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            Step Over
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={onStop}>
              <Square className="h-4 w-4 fill-current" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            Stop
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}
