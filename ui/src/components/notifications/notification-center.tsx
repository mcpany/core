/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Bell, Check, ExternalLink, Info, AlertTriangle, AlertCircle } from "lucide-react";
import Link from "next/link";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/lib/client";
import { Alert } from "@/components/alerts/types";
import { formatDistanceToNow } from "date-fns";

/**
 * NotificationCenter component.
 * Displays a bell icon with a badge for active alerts, and a dropdown list of recent alerts.
 */
export function NotificationCenter() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [isOpen, setIsOpen] = useState(false);

  const fetchAlerts = async () => {
    try {
      const data = await apiClient.listAlerts();
      // Filter for active alerts
      const active = data.filter((a: any) => a.status === 'active');
      // Sort by timestamp desc
      active.sort((a: any, b: any) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
      setAlerts(active);
    } catch (e) {
      console.error("Failed to fetch alerts", e);
    }
  };

  useEffect(() => {
    fetchAlerts();
    // Poll every 30 seconds
    const interval = setInterval(fetchAlerts, 30000);
    return () => clearInterval(interval);
  }, []);

  const markAsRead = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    try {
      await apiClient.updateAlertStatus(id, 'acknowledged');
      // Optimistic update
      setAlerts(prev => prev.filter(a => a.id !== id));
    } catch (e) {
      console.error("Failed to update alert", e);
    }
  };

  const getIcon = (severity: string) => {
    switch (severity) {
      case 'critical': return <AlertCircle className="h-4 w-4 text-red-500" />;
      case 'warning': return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      default: return <Info className="h-4 w-4 text-blue-500" />;
    }
  };

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative" aria-label="Notifications">
          <Bell className="h-4 w-4" />
          {alerts.length > 0 && (
            <span className="absolute top-2 right-2 flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="flex items-center justify-between p-4 border-b">
          <h4 className="font-semibold leading-none">Notifications</h4>
          {alerts.length > 0 && (
             <Badge variant="secondary" className="text-xs">{alerts.length} New</Badge>
          )}
        </div>
        <ScrollArea className="h-[300px]">
          {alerts.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-32 text-muted-foreground">
              <Bell className="h-8 w-8 mb-2 opacity-20" />
              <p className="text-sm">No new notifications</p>
            </div>
          ) : (
            <div className="flex flex-col divide-y">
              {alerts.map(alert => (
                <div key={alert.id} className="p-4 hover:bg-muted/50 transition-colors relative group">
                  <div className="flex gap-3 items-start">
                    <div className="mt-0.5 shrink-0">
                      {getIcon(alert.severity)}
                    </div>
                    <div className="flex-1 space-y-1">
                      <p className="text-sm font-medium leading-none pr-6">
                        {alert.title}
                      </p>
                      <p className="text-xs text-muted-foreground line-clamp-2">
                        {alert.message}
                      </p>
                      <div className="flex items-center gap-2 mt-1.5">
                        <Badge variant="outline" className="text-[10px] h-4 px-1">{alert.service}</Badge>
                        <span className="text-[10px] text-muted-foreground">
                          {formatDistanceToNow(new Date(alert.timestamp), { addSuffix: true })}
                        </span>
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6 absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={(e) => markAsRead(e, alert.id)}
                      title="Mark as read"
                    >
                      <Check className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
        <div className="p-2 border-t bg-muted/10">
          <Link href="/alerts" passHref onClick={() => setIsOpen(false)}>
            <Button variant="ghost" size="sm" className="w-full text-xs h-8">
              View All Alerts <ExternalLink className="ml-2 h-3 w-3" />
            </Button>
          </Link>
        </div>
      </PopoverContent>
    </Popover>
  );
}
