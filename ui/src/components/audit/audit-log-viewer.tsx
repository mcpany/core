import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { RichResultViewer } from '@/components/tools/rich-result-viewer';
import { JsonView } from '@/components/ui/json-view';
import { Alert, AlertDescription } from '@/components/ui/alert';

export function AuditLogViewer() {
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchLogs() {
      try {
        const response = await fetch('/api/v1/debug/traces');
        if (!response.ok) {
          throw new Error('Failed to fetch audit logs');
        }
        const data = await response.json();
        // Extract array if it's wrapped in an object
        const logArray = Array.isArray(data) ? data : data.entries || data.logs || [];
        setLogs(logArray);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    fetchLogs();
  }, []);

  if (loading) {
    return <div className="p-4 text-muted-foreground">Loading audit logs...</div>;
  }

  if (error) {
    return (
      <Alert variant="destructive" className="m-4">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-4 p-4">
      <Card className="border shadow-sm bg-background/50 backdrop-blur-sm">
        <CardHeader className="border-b bg-muted/20">
          <CardTitle className="text-lg font-medium">Audit Trail</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {logs.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">
              No audit logs found.
            </div>
          ) : (
            <div className="p-4">
              <RichResultViewer
                result={JSON.stringify(logs)}
                defaultView="table"
              />
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border shadow-sm">
        <CardHeader className="border-b bg-muted/20">
          <CardTitle className="text-lg font-medium">Raw JSON</CardTitle>
        </CardHeader>
        <CardContent className="p-4">
          <JsonView src={logs} collapsed={1} />
        </CardContent>
      </Card>
    </div>
  );
}
