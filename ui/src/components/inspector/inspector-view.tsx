/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Button } from "@/components/ui/button";
import { useInspectorStream, MCPMessage } from "@/hooks/use-inspector-stream";
import { MessageList } from "./message-list";
import { MessageDetail } from "./message-detail";
import { Play, Pause, Trash2, Activity, Zap } from "lucide-react";
import { Badge } from "@/components/ui/badge";

export function InspectorView() {
    const {
        messages,
        isConnected,
        isPaused,
        isSimulating,
        clearMessages,
        togglePause,
        startSimulation,
        stopSimulation
    } = useInspectorStream();

    const [selectedMessage, setSelectedMessage] = useState<MCPMessage | null>(null);

    return (
        <div className="flex flex-col h-full bg-background">
            {/* Toolbar */}
            <div className="flex items-center justify-between p-4 border-b bg-muted/20">
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                        <Activity className="h-5 w-5 text-primary" />
                        <h2 className="text-lg font-semibold">Inspector</h2>
                    </div>
                    <div className="flex items-center gap-2">
                        <Badge variant={isConnected ? "default" : "secondary"} className="h-6">
                            {isConnected ? "Connected" : "Disconnected"}
                        </Badge>
                        {isSimulating && (
                            <Badge variant="outline" className="h-6 text-amber-500 border-amber-500/50">
                                Simulation Mode
                            </Badge>
                        )}
                        <Badge variant="outline" className="h-6">
                            {messages.length} Messages
                        </Badge>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                     <Button
                        variant="outline"
                        size="sm"
                        onClick={isSimulating ? stopSimulation : startSimulation}
                        className={isSimulating ? "bg-amber-500/10 text-amber-500 hover:bg-amber-500/20" : ""}
                    >
                        <Zap className="mr-2 h-4 w-4" />
                        {isSimulating ? "Stop Simulation" : "Simulate Traffic"}
                    </Button>
                    <Button variant="outline" size="sm" onClick={togglePause}>
                        {isPaused ? <Play className="mr-2 h-4 w-4" /> : <Pause className="mr-2 h-4 w-4" />}
                        {isPaused ? "Resume" : "Pause"}
                    </Button>
                    <Button variant="outline" size="sm" onClick={clearMessages}>
                        <Trash2 className="mr-2 h-4 w-4" /> Clear
                    </Button>
                </div>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-hidden">
                <ResizablePanelGroup direction="horizontal">
                    <ResizablePanel defaultSize={40} minSize={30}>
                        <MessageList
                            messages={messages}
                            selectedId={selectedMessage?.id || null}
                            onSelect={setSelectedMessage}
                        />
                    </ResizablePanel>

                    <ResizableHandle />

                    <ResizablePanel defaultSize={60}>
                        <MessageDetail
                            message={selectedMessage}
                            onClose={() => setSelectedMessage(null)}
                        />
                    </ResizablePanel>
                </ResizablePanelGroup>
            </div>
        </div>
    );
}
