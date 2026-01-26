/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import {
  Zap,
  Server,
  Database,
  Users,
  Webhook,
  Layers,
  MessageSquare,
  Link as LinkIcon,
  Cpu,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface LegendItemProps {
    icon: React.ReactNode;
    label: string;
    description: string;
    color: string;
}

/**
 * LegendItem component.
 * @param props - The component props.
 * @param props.icon - The icon property.
 * @param props.label - The label property.
 * @param props.description - The description property.
 * @param props.color - The color property.
 * @returns The rendered component.
 */
const LegendItem = React.memo(({ icon, label, description, color }: LegendItemProps) => (
    <div className="flex items-start gap-3 p-2 rounded-md hover:bg-muted/50 transition-colors">
        <div className={cn("p-1.5 rounded-md", color)}>
            {React.cloneElement(icon as React.ReactElement, { className: "h-4 w-4 text-white" })}
        </div>
        <div className="space-y-0.5">
            <h4 className="text-sm font-medium leading-none">{label}</h4>
            <p className="text-xs text-muted-foreground">{description}</p>
        </div>
    </div>
));
LegendItem.displayName = 'LegendItem';

/**
 * NetworkLegend component.
 * Displays a legend for the network topology graph, explaining node types and status indicators.
 *
 * @returns The rendered NetworkLegend component.
 */
export const NetworkLegend = React.memo(function NetworkLegend() {
    return (
        <div className="space-y-1">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                 <LegendItem
                    icon={<Cpu />}
                    label="Core"
                    description="The main MCP Any server instance."
                    color="bg-blue-500"
                />
                <LegendItem
                    icon={<Server />}
                    label="Service"
                    description="Upstream API or backend service."
                    color="bg-indigo-500"
                />
                 <LegendItem
                    icon={<Users />}
                    label="Client"
                    description="Connected MCP Client (e.g. Claude)."
                    color="bg-green-500"
                />
                 <LegendItem
                    icon={<Zap />}
                    label="Tool"
                    description="Executable capability exposed to the AI."
                    color="bg-amber-500"
                />
                <LegendItem
                    icon={<Layers />}
                    label="Middleware"
                    description="Interceptors like Auth or Logging."
                    color="bg-orange-500"
                />
                <LegendItem
                    icon={<Database />}
                    label="Resource"
                    description="Data source or file exposed to the AI."
                    color="bg-cyan-500"
                />
                 <LegendItem
                    icon={<MessageSquare />}
                    label="Prompt"
                    description="Pre-defined prompt template."
                    color="bg-purple-500"
                />
                 <LegendItem
                    icon={<LinkIcon />}
                    label="API Call"
                    description="Specific API endpoint call representation."
                    color="bg-slate-500"
                />
                <LegendItem
                    icon={<Webhook />}
                    label="Webhook"
                    description="Event-driven webhook receiver."
                    color="bg-pink-500"
                />
            </div>

            <div className="pt-4 mt-4 border-t">
                <h4 className="text-xs font-semibold uppercase text-muted-foreground mb-3 tracking-wider">Status Indicators</h4>
                <div className="flex gap-4 text-xs">
                     <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]"></div>
                        <span>Active</span>
                    </div>
                     <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-slate-400 border border-slate-600"></div>
                        <span>Inactive</span>
                    </div>
                     <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)] animate-pulse"></div>
                        <span>Error</span>
                    </div>
                </div>
            </div>
        </div>
    );
});
NetworkLegend.displayName = 'NetworkLegend';
