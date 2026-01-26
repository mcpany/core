"use client"

import { useState, useEffect } from "react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Check, Copy, Download, Terminal } from "lucide-react"

export default function IntegrationsPage() {
  const [origin, setOrigin] = useState("http://localhost:50050")
  const [copied, setCopied] = useState<string | null>(null)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      // Try to intelligently detect the server URL
      const url = new URL(window.location.href)
      // If we are on the UI dev port (9002), assume server is on 50050
      if (url.port === '9002') {
          setOrigin('http://localhost:50050')
      } else {
          // Otherwise assume we are serving from the same host/port (production/docker)
          setOrigin(url.origin)
      }
    }
  }, [])

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopied(id)
    setTimeout(() => setCopied(null), 2000)
  }

  const downloadConfig = () => {
    const config = getClaudeConfig(origin);
    const blob = new Blob([config], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "claude_desktop_config.json";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  const getClaudeConfig = (url: string) => JSON.stringify({
    "mcpServers": {
      "mcp-any": {
        "command": "npx",
        "args": [
          "-y",
          "mcp-proxy",
          `${url}/sse`
        ]
      }
    }
  }, null, 2)

  const claudeConfig = getClaudeConfig(origin);

  const cursorConfig = `// Add this to your Cursor MCP settings
// Usually found in Settings > Features > MCP

Name: MCP Any
Type: SSE
URL:  ${origin}/sse`

  const vsCodeConfig = JSON.stringify({
    "mcp.servers": {
        "mcp-any": {
            "url": `${origin}/sse`,
            "transport": "sse"
        }
    }
  }, null, 2)

  const geminiCommand = `gemini mcp add --transport http --trust mcpany ${origin}`

  return (
      <div className="space-y-6 p-6">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">Integrations</h1>
            <p className="text-muted-foreground">Connect your favorite AI clients to MCP Any.</p>
        </div>

        <Tabs defaultValue="claude" className="w-full">
            <TabsList className="grid w-full grid-cols-4 lg:w-[600px]">
                <TabsTrigger value="claude">Claude Desktop</TabsTrigger>
                <TabsTrigger value="cursor">Cursor</TabsTrigger>
                <TabsTrigger value="vscode">VS Code</TabsTrigger>
                <TabsTrigger value="gemini">Gemini</TabsTrigger>
            </TabsList>

            <TabsContent value="claude">
                <Card>
                    <CardHeader>
                        <CardTitle>Claude Desktop Configuration</CardTitle>
                        <CardDescription>
                            Add this configuration to your <code>claude_desktop_config.json</code> file.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="rounded-md bg-muted p-4 relative group">
                            <pre className="text-sm font-mono overflow-auto">{claudeConfig}</pre>
                            <Button
                                size="icon"
                                variant="ghost"
                                className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
                                onClick={() => copyToClipboard(claudeConfig, 'claude')}
                            >
                                {copied === 'claude' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                            </Button>
                        </div>
                        <div className="flex gap-2">
                            <Button onClick={() => copyToClipboard(claudeConfig, 'claude')}>
                                {copied === 'claude' ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
                                Copy Config
                            </Button>
                            <Button variant="secondary" onClick={downloadConfig}>
                                <Download className="mr-2 h-4 w-4" />
                                Download Config File
                            </Button>
                        </div>
                        <div className="text-sm text-muted-foreground">
                            <p className="font-medium">File locations:</p>
                            <ul className="list-disc list-inside ml-2 mt-1 space-y-1">
                                <li>macOS: <code>~/Library/Application Support/Claude/claude_desktop_config.json</code></li>
                                <li>Windows: <code>%APPDATA%\Claude\claude_desktop_config.json</code></li>
                            </ul>
                        </div>
                    </CardContent>
                </Card>
            </TabsContent>

            <TabsContent value="cursor">
                <Card>
                    <CardHeader>
                        <CardTitle>Cursor Configuration</CardTitle>
                        <CardDescription>
                            Add a new MCP server in Cursor Settings.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="rounded-md bg-muted p-4 relative group">
                            <pre className="text-sm font-mono overflow-auto">{cursorConfig}</pre>
                            <Button
                                size="icon"
                                variant="ghost"
                                className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
                                onClick={() => copyToClipboard(cursorConfig, 'cursor')}
                            >
                                {copied === 'cursor' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                            </Button>
                        </div>
                        <Button onClick={() => copyToClipboard(`${origin}/sse`, 'cursor-url')}>
                            {copied === 'cursor-url' ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
                            Copy URL Only
                        </Button>
                    </CardContent>
                </Card>
            </TabsContent>

            <TabsContent value="vscode">
                <Card>
                    <CardHeader>
                        <CardTitle>VS Code Configuration</CardTitle>
                        <CardDescription>
                            Use this snippet for generic MCP extensions in VS Code settings.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="rounded-md bg-muted p-4 relative group">
                            <pre className="text-sm font-mono overflow-auto">{vsCodeConfig}</pre>
                            <Button
                                size="icon"
                                variant="ghost"
                                className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
                                onClick={() => copyToClipboard(vsCodeConfig, 'vscode')}
                            >
                                {copied === 'vscode' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                            </Button>
                        </div>
                        <Button onClick={() => copyToClipboard(vsCodeConfig, 'vscode')}>
                            {copied === 'vscode' ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
                            Copy Config
                        </Button>
                    </CardContent>
                </Card>
            </TabsContent>

            <TabsContent value="gemini">
                <Card>
                    <CardHeader>
                        <CardTitle>Gemini CLI</CardTitle>
                        <CardDescription>
                            Run this command in your terminal to connect Gemini.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="rounded-md bg-muted p-4 relative group flex items-center gap-2">
                            <Terminal className="h-4 w-4 text-muted-foreground shrink-0" />
                            <pre className="text-sm font-mono overflow-auto flex-1">{geminiCommand}</pre>
                            <Button
                                size="icon"
                                variant="ghost"
                                className="opacity-0 group-hover:opacity-100 transition-opacity"
                                onClick={() => copyToClipboard(geminiCommand, 'gemini')}
                            >
                                {copied === 'gemini' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                            </Button>
                        </div>
                        <Button onClick={() => copyToClipboard(geminiCommand, 'gemini')}>
                            {copied === 'gemini' ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
                            Copy Command
                        </Button>
                    </CardContent>
                </Card>
            </TabsContent>
        </Tabs>
      </div>
  )
}
