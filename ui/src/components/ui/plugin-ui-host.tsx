/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Loader2, ExternalLink, ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface PluginUIHostProps {
  src: string;
  title?: string;
  className?: string;
  serviceId?: string;
}

/**
 * Component to host custom UI provided by server plugins via iframe.
 * Includes security sandboxing and loading states.
 *
 * @param serviceId } - PluginUIHostProps. Description.
 */
export function PluginUIHost({ src, title = "Plugin UI", className, serviceId }: PluginUIHostProps) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handleLoad = () => {
    setLoading(false);
  };

  const handleError = () => {
    setLoading(false);
    setError("Failed to load plugin UI. This could be due to a connection issue or security policy.");
  };

  if (!src) {
    return (
        <div className="flex flex-col items-center justify-center h-[300px] border-2 border-dashed rounded-lg bg-muted/20">
            <ShieldAlert className="h-8 w-8 text-muted-foreground mb-2" />
            <p className="text-sm text-muted-foreground">No UI URL provided by plugin.</p>
        </div>
    );
  }

  return (
    <div className={cn("relative w-full h-full min-h-[400px] border rounded-lg overflow-hidden bg-background", className)}>
      {loading && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-background/80 z-10">
          <Loader2 className="h-8 w-8 animate-spin text-primary mb-2" />
          <p className="text-sm text-muted-foreground italic">Launching plugin interface...</p>
        </div>
      )}

      {error ? (
          <div className="absolute inset-0 flex flex-col items-center justify-center p-6 text-center space-y-4">
              <ShieldAlert className="h-10 w-10 text-destructive mb-2" />
              <h3 className="font-semibold">Integration Error</h3>
              <p className="text-sm text-muted-foreground max-w-sm">{error}</p>
              <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => window.open(src, '_blank')}>
                      <ExternalLink className="h-4 w-4 mr-2" /> Open in New Tab
                  </Button>
                  <Button variant="ghost" size="sm" onClick={() => { setError(null); setLoading(true); }}>
                      Try Again
                  </Button>
              </div>
          </div>
      ) : (
        <iframe
            src={src}
            title={title}
            className="w-full h-full border-none"
            onLoad={handleLoad}
            onError={handleError}
            // SECURITY WARNING: allow-same-origin is dangerous if the plugin content is not trusted.
            // It allows the iframe to access the same local storage and cookies as the parent if origins match.
            // Ensure plugins are from trusted sources or served from a different origin.
            sandbox="allow-scripts allow-forms allow-popups allow-same-origin"
            referrerPolicy="no-referrer"
        />
      )}

      {serviceId && !loading && !error && (
        <div className="absolute bottom-2 right-2 opacity-50 hover:opacity-100 transition-opacity">
            <div className="text-[10px] bg-muted/80 px-1.5 py-0.5 rounded border">
                Hosted by: <strong>{serviceId}</strong>
            </div>
        </div>
      )}
    </div>
  );
}
