// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

'use client';

import React, { useEffect, useState, useMemo } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skill, SkillService } from '@/lib/skill-service';
import { Edit, ChevronLeft, AlertTriangle, Copy, Terminal, CheckCircle2, Sparkles } from 'lucide-react';
import { toast } from 'sonner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { apiClient, ToolDefinition } from '@/lib/client';
import { JsonView } from '@/components/ui/json-view';

/**
 * SkillDetail component.
 * @returns The rendered component.
 */
export default function SkillDetail() {
  const params = useParams();
  const name = params?.name as string | undefined;

  const [skill, setSkill] = useState<Skill | null>(null);
  const [availableTools, setAvailableTools] = useState<ToolDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (name) loadData(name);
  }, [name]);

  const loadData = async (skillName: string) => {
    try {
      setLoading(true);
      const [skillData, toolsData] = await Promise.all([
          SkillService.get(skillName),
          apiClient.listTools().catch(() => ({ tools: [] }))
      ]);
      setSkill(skillData);
      setAvailableTools(toolsData?.tools || []);
    } catch (err: any) {
      toast.error('Failed to load data: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const missingTools = useMemo(() => {
      if (!skill || !skill.allowedTools) return [];
      return skill.allowedTools.filter(t => !availableTools.find(at => at.name === t));
  }, [skill, availableTools]);

  const resolvedTools = useMemo(() => {
      if (!skill || !skill.allowedTools) return [];
      return availableTools.filter(t => skill.allowedTools?.includes(t.name));
  }, [skill, availableTools]);

  const systemContext = useMemo(() => {
      if (!skill) return "";
      const toolsJson = JSON.stringify(resolvedTools.map(t => ({
          name: t.name,
          description: t.description,
          inputSchema: t.inputSchema
      })), null, 2);

      return `# Instructions
${skill.instructions}

# Available Tools
${toolsJson}`;
  }, [skill, resolvedTools]);

  const handleCopyContext = () => {
      navigator.clipboard.writeText(systemContext);
      setCopied(true);
      toast.success("System context copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
  };

  if (loading) return <div>Loading skill...</div>;
  if (!skill) return <div>Skill not found</div>;

  return (
    <div className="container mx-auto py-8 max-w-4xl">
      <div className="mb-6">
        <Link href="/skills">
          <Button variant="ghost" className="pl-0">
            <ChevronLeft className="mr-2 h-4 w-4" /> Back to Skills
          </Button>
        </Link>
      </div>

      <div className="flex justify-between items-start mb-6">
        <div>
          <h1 className="text-4xl font-bold mb-2">{skill.name}</h1>
          <p className="text-xl text-muted-foreground">{skill.description}</p>
        </div>
        <Link href={`/skills/${skill.name}/edit`}>
          <Button variant="outline">
            <Edit className="mr-2 h-4 w-4" /> Edit Skill
          </Button>
        </Link>
      </div>

      <Tabs defaultValue="overview" className="w-full">
          <TabsList className="grid w-full grid-cols-2 mb-8">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="simulation">Simulation & Workbench</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Metadata</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                 {skill.license && (
                     <div>
                         <span className="font-semibold mr-2">License:</span>
                         {skill.license}
                     </div>
                 )}
                 <div>
                    <span className="font-semibold block mb-2">Allowed Tools:</span>
                    <div className="flex gap-2 flex-wrap">
                        {skill.allowedTools && skill.allowedTools.length > 0 ? (
                            skill.allowedTools.map(t => <Badge key={t} variant="secondary">{t}</Badge>)
                        ) : (
                            <span className="text-muted-foreground italic">None allowed (default)</span>
                        )}
                    </div>
                 </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Instructions</CardTitle>
              </CardHeader>
              <CardContent>
                <pre className="bg-muted p-4 rounded-lg overflow-x-auto whitespace-pre-wrap font-mono text-sm">
                    {skill.instructions}
                </pre>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Assets</CardTitle>
              </CardHeader>
              <CardContent>
                  {skill.assets && skill.assets.length > 0 ? (
                      <ul className="list-disc pl-5">
                          {skill.assets.map(a => <li key={a} className="font-mono">{a}</li>)}
                      </ul>
                  ) : (
                      <p className="text-muted-foreground">No assets.</p>
                  )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="simulation" className="space-y-6">
              {/* Validation Status */}
              {missingTools.length > 0 ? (
                  <Alert variant="destructive">
                      <AlertTriangle className="h-4 w-4" />
                      <AlertTitle>Missing Tools Detected</AlertTitle>
                      <AlertDescription>
                          The following tools are required by this skill but were not found in the current registry:
                          <div className="mt-2 font-mono font-bold">
                              {missingTools.join(', ')}
                          </div>
                      </AlertDescription>
                  </Alert>
              ) : (
                  <Alert className="bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:border-green-900 dark:text-green-300">
                      <CheckCircle2 className="h-4 w-4" />
                      <AlertTitle>All Tools Available</AlertTitle>
                      <AlertDescription>
                          All {skill.allowedTools?.length || 0} required tools are connected and available.
                      </AlertDescription>
                  </Alert>
              )}

              {/* Context Preview */}
              <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <div className="space-y-1">
                          <CardTitle className="text-base font-medium flex items-center gap-2">
                              <Sparkles className="h-4 w-4 text-primary" />
                              System Context Preview
                          </CardTitle>
                          <CardDescription>
                              This is the exact context (Instructions + Tool Definitions) that will be sent to the LLM.
                          </CardDescription>
                      </div>
                      <Button variant="outline" size="sm" onClick={handleCopyContext}>
                          {copied ? <CheckCircle2 className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
                          {copied ? "Copied" : "Copy to Clipboard"}
                      </Button>
                  </CardHeader>
                  <CardContent className="pt-4">
                      <div className="relative">
                          <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-lg overflow-x-auto font-mono text-xs h-[400px] border">
                              {systemContext}
                          </pre>
                          <div className="absolute top-2 right-2 text-[10px] text-muted-foreground bg-black/50 px-2 py-1 rounded">
                              {systemContext.length} chars
                          </div>
                      </div>
                  </CardContent>
              </Card>

              {/* Tool Explorer */}
              <div className="space-y-4">
                  <h3 className="text-lg font-medium flex items-center gap-2">
                      <Terminal className="h-5 w-5" /> Resolved Tool Schemas
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {resolvedTools.map(tool => (
                          <Card key={tool.name} className="overflow-hidden">
                              <CardHeader className="py-3 bg-muted/30 border-b">
                                  <CardTitle className="text-sm font-mono">{tool.name}</CardTitle>
                                  <CardDescription className="line-clamp-1 text-xs">{tool.description}</CardDescription>
                              </CardHeader>
                              <CardContent className="p-0">
                                  <JsonView data={tool.inputSchema} maxHeight={200} />
                              </CardContent>
                          </Card>
                      ))}
                      {resolvedTools.length === 0 && (
                          <div className="col-span-full p-8 text-center text-muted-foreground border rounded-lg border-dashed">
                              No tools selected or available.
                          </div>
                      )}
                  </div>
              </div>
          </TabsContent>
      </Tabs>
    </div>
  );
}
