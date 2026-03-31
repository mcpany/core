"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

export function InteropTester() {
  const [framework, setFramework] = useState("CrewAI");
  const [intent, setIntent] = useState("task_delegation");
  const [payloadStr, setPayloadStr] = useState('{"role": "data_analyst"}');
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleTest = async () => {
    setLoading(true);
    setResult(null);
    try {
      let payload = {};
      try {
        payload = JSON.parse(payloadStr);
      } catch (e) {
        toast({ title: "Invalid JSON", description: "Payload must be valid JSON", variant: "destructive" });
        return;
      }

      const task = {
        id: "ui-test-" + Date.now(),
        framework,
        intent,
        payload,
      };

      const res = await apiClient.postInteropTask(task);
      setResult(JSON.stringify(res, null, 2));
      toast({ title: "Task Executed", description: "Successfully routed task" });
    } catch (err: any) {
      toast({ title: "Error", description: err.message, variant: "destructive" });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Interop Task Tester</CardTitle>
        <CardDescription>Send a task to an agent framework via the Adapter Hub</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="text-sm font-medium">Framework</label>
          <Input value={framework} onChange={(e) => setFramework(e.target.value)} placeholder="e.g., CrewAI, OpenClaw, AutoGen" />
        </div>
        <div>
          <label className="text-sm font-medium">Intent</label>
          <Input value={intent} onChange={(e) => setIntent(e.target.value)} placeholder="e.g., task_delegation" />
        </div>
        <div>
          <label className="text-sm font-medium">Payload (JSON)</label>
          <Input value={payloadStr} onChange={(e) => setPayloadStr(e.target.value)} />
        </div>
        <Button onClick={handleTest} disabled={loading}>
          {loading ? "Sending..." : "Send Task"}
        </Button>

        {result && (
          <div className="mt-4 p-4 bg-muted rounded-md overflow-auto max-h-60 text-xs">
            <pre data-testid="interop-result">{result}</pre>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
