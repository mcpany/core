// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

'use client';

import React, { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skill, SkillService } from '@/lib/skill-service';
import { Edit, ChevronLeft } from 'lucide-react';
import { toast } from 'sonner';

/**
 * SkillDetail component.
 * @returns The rendered component.
 */
export default function SkillDetail() {
  const params = useParams();
  const name = params?.name as string | undefined;

  const [skill, setSkill] = useState<Skill | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (name) loadSkill(name);
  }, [name]);

  const loadSkill = async (skillName: string) => {
    try {
      setLoading(true);
      const data = await SkillService.get(skillName);
      setSkill(data);
    } catch (err: any) {
      toast.error('Failed to load skill: ' + err.message);
    } finally {
      setLoading(false);
    }
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

      <div className="grid gap-6">
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
      </div>
    </div>
  );
}
