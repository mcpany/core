/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { JsonView } from "@/components/ui/json-view";
import { Link as LinkIcon, Check, Copy } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * ConnectClientButton component.
 * Provides a modal with configuration snippets for connecting various MCP clients.
 *
 * @returns {JSX.Element} The rendered component.
 */
export function ConnectClientButton() {
  const [isOpen, setIsOpen] = useState(false);
  const [origin, setOrigin] = useState("");
  const [apiKey, setApiKey] = useState("");
  const { toast } = useToast();

  useEffect(() => {
    if (typeof window !== "undefined") {
      setOrigin(window.location.origin);
    }
  }, []);

  const sseUrl = `${origin}/sse`;
  const displayUrl = origin || "http://localhost:50050";

  // Construct URL with API Key if present (for URL-based auth clients)
  const authenticatedUrl = apiKey ? `${displayUrl}/sse?api_key=${apiKey}` : `${displayUrl}/sse`;

  const copyToClipboard = (text: string) => {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).catch(console.error);
        toast({ title: "Copied", description: "Copied to clipboard" });
    }
  };

  // Configurations
  const cursorConfig = {
      name: "MCP Any",
      type: "sse",
      url: authenticatedUrl
  };

  const claudeConfig = {
      "mcpServers": {
          "mcp-any": {
              "command": "npx",
              "args": [
                  "-y",
                  "@modelcontextprotocol/server-sse-client",
                  "--url",
                  authenticatedUrl
              ]
          }
      }
  };

  const vsCodeConfig = {
      "mcp.servers": {
          "mcp-any": {
              "name": "MCP Any",
              "url": authenticatedUrl
          }
      }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2 hidden md:flex">
          <LinkIcon className="h-4 w-4" />
          Connect
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Connect to MCP Any</DialogTitle>
          <DialogDescription>
            Configure your AI client to access tools managed by MCP Any.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="api-key" className="text-right">
                    API Key
                </Label>
                <Input
                    id="api-key"
                    placeholder="Optional (if configured)"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    className="col-span-3"
                />
            </div>

            <Tabs defaultValue="claude" className="w-full">
                <TabsList className="grid w-full grid-cols-4">
                    <TabsTrigger value="claude">Claude</TabsTrigger>
                    <TabsTrigger value="cursor">Cursor</TabsTrigger>
                    <TabsTrigger value="vscode">VS Code</TabsTrigger>
                    <TabsTrigger value="cli">CLI</TabsTrigger>
                </TabsList>

                <TabsContent value="claude" className="space-y-4 pt-4">
                    <div className="space-y-2">
                        <h4 className="font-medium text-sm">Claude Desktop Configuration</h4>
                        <p className="text-xs text-muted-foreground">
                            Edit your <code>claude_desktop_config.json</code> and add this entry.
                            Note: Requires Node.js installed to run the SSE client bridge.
                        </p>
                    </div>
                    <JsonView data={claudeConfig} />
                    <div className="text-xs text-muted-foreground bg-muted p-2 rounded border">
                        <p className="font-semibold mb-1">Config Locations:</p>
                        <ul className="list-disc pl-4 space-y-1">
                            <li><strong>macOS:</strong> <code>~/Library/Application Support/Claude/claude_desktop_config.json</code></li>
                            <li><strong>Windows:</strong> <code>%APPDATA%\Claude\claude_desktop_config.json</code></li>
                        </ul>
                    </div>
                </TabsContent>

                <TabsContent value="cursor" className="space-y-4 pt-4">
                    <div className="space-y-2">
                        <h4 className="font-medium text-sm">Cursor Setup</h4>
                        <ol className="list-decimal pl-4 text-sm space-y-2 text-muted-foreground">
                            <li>Open Cursor Settings (Cmd+,)</li>
                            <li>Navigate to <strong>Features &gt; MCP</strong></li>
                            <li>Click <strong>+ Add New MCP Server</strong></li>
                            <li>Select Type: <strong>SSE</strong></li>
                            <li>Enter Name: <code>MCP Any</code></li>
                            <li>Enter URL below:</li>
                        </ol>
                    </div>
                    <div className="flex items-center space-x-2">
                        <Input readOnly value={authenticatedUrl} className="font-mono text-xs" />
                        <Button size="icon" variant="outline" onClick={() => copyToClipboard(authenticatedUrl)}>
                            <Copy className="h-4 w-4" />
                        </Button>
                    </div>
                </TabsContent>

                <TabsContent value="vscode" className="space-y-4 pt-4">
                    <div className="space-y-2">
                         <h4 className="font-medium text-sm">VS Code (Generic MCP Extension)</h4>
                         <p className="text-xs text-muted-foreground">
                            Add this to your VS Code <code>settings.json</code> if you are using an MCP extension that supports generic servers.
                        </p>
                    </div>
                    <JsonView data={vsCodeConfig} />
                </TabsContent>

                <TabsContent value="cli" className="space-y-4 pt-4">
                    <div className="space-y-2">
                         <h4 className="font-medium text-sm">Gemini CLI</h4>
                         <p className="text-xs text-muted-foreground">Run this command in your terminal:</p>
                    </div>
                    <div className="relative group">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background"
                            onClick={() => copyToClipboard(`gemini mcp add --transport http --trust mcpany ${displayUrl}`)}
                        >
                            <Copy className="h-3 w-3" />
                        </Button>
                        <pre className="text-xs font-mono bg-muted/50 p-3 rounded-md overflow-x-auto border">
                            gemini mcp add --transport http --trust mcpany {displayUrl}
                        </pre>
                    </div>
                </TabsContent>
            </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  );
}
