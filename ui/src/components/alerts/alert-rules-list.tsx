/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Trash2, Loader2, RefreshCw } from "lucide-react";
import { apiClient, AlertRule } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { formatDistanceToNow } from "date-fns";

interface AlertRulesListProps {
  refreshTrigger?: number;
}

export function AlertRulesList({ refreshTrigger }: AlertRulesListProps) {
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchRules = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listAlertRules();
      setRules(data || []);
    } catch (error) {
      console.error(error);
      toast({
        title: "Error",
        description: "Failed to load alert rules",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRules();
  }, [refreshTrigger]);

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this rule?")) return;
    try {
      await apiClient.deleteAlertRule(id);
      toast({ title: "Rule Deleted", description: "Alert rule removed successfully." });
      fetchRules();
    } catch (error) {
      console.error(error);
      toast({
        title: "Error",
        description: "Failed to delete rule",
        variant: "destructive",
      });
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case "critical": return <Badge variant="destructive" className="uppercase text-[10px]">Critical</Badge>;
      case "warning": return <Badge variant="secondary" className="bg-yellow-500/15 text-yellow-600 dark:text-yellow-400 uppercase text-[10px]">Warning</Badge>;
      case "info": return <Badge variant="outline" className="text-blue-500 border-blue-200 dark:border-blue-800 uppercase text-[10px]">Info</Badge>;
      default: return <Badge variant="outline" className="uppercase text-[10px]">{severity}</Badge>;
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button variant="ghost" size="sm" onClick={fetchRules} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>
      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Severity</TableHead>
              <TableHead>Condition</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rules.length === 0 ? (
                 <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                        <div className="flex items-center justify-center gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Loading rules...
                        </div>
                    </TableCell>
                </TableRow>
            ) : rules.length === 0 ? (
                <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                        No alert rules defined. Create one to start monitoring.
                    </TableCell>
                </TableRow>
            ) : (
                rules.map((rule) => (
                <TableRow key={rule.id}>
                    <TableCell className="font-medium">{rule.name}</TableCell>
                    <TableCell>{getSeverityBadge(rule.severity)}</TableCell>
                    <TableCell className="font-mono text-xs">
                        {rule.metric} {rule.operator} {rule.threshold} (for {rule.duration})
                    </TableCell>
                    <TableCell>
                        <Badge variant={rule.enabled ? "default" : "secondary"}>
                            {rule.enabled ? "Enabled" : "Disabled"}
                        </Badge>
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                        {rule.last_updated ? formatDistanceToNow(new Date(rule.last_updated), { addSuffix: true }) : "-"}
                    </TableCell>
                    <TableCell className="text-right">
                        <Button variant="ghost" size="icon" onClick={() => rule.id && handleDelete(rule.id)}>
                            <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
                        </Button>
                    </TableCell>
                </TableRow>
                ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
