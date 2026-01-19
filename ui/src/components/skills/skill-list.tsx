// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

'use client';

import React, { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardDescription, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Plus, Trash2, Edit } from 'lucide-react';
import { Skill, SkillService } from '@/lib/skill-service';
import Link from 'next/link';
import { toast } from 'sonner';

/**
 * SkillList component.
 * @returns The rendered component.
 */
export default function SkillList() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSkills();
  }, []);

  const loadSkills = async () => {
    try {
      const list = await SkillService.list();
      setSkills(list);
    } catch (err: any) {
      toast.error('Failed to load skills: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Are you sure you want to delete the skill "${name}"?`)) return;
    try {
      await SkillService.delete(name);
      toast.success('Skill deleted');
      loadSkills();
    } catch (err: any) {
      toast.error('Failed to delete skill: ' + err.message);
    }
  };

  if (loading) {
    return <div className="p-4">Loading skills...</div>;
  }

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Agent Skills</h1>
        <Link href="/skills/create">
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Create Skill
          </Button>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {skills.map((skill) => (
          <Card key={skill.name} className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle className="flex justify-between items-center">
                <span>{skill.name}</span>
                <div className="flex gap-2">
                  <Link href={`/skills/${skill.name}/edit`}>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <Edit className="h-4 w-4" />
                    </Button>
                  </Link>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-destructive hover:text-destructive"
                    onClick={() => handleDelete(skill.name)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </CardTitle>
              <CardDescription className="line-clamp-2 h-10">
                {skill.description}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                {skill.allowedTools && skill.allowedTools.length > 0 && (
                  <span className="bg-secondary px-2 py-1 rounded">
                    {skill.allowedTools.length} Tools
                  </span>
                )}
                {skill.assets && skill.assets.length > 0 && (
                  <span className="bg-secondary px-2 py-1 rounded">
                    {skill.assets.length} Assets
                  </span>
                )}
                <Link href={`/skills/${skill.name}`} className="ml-auto text-primary hover:underline">
                  View Details
                </Link>
              </div>
            </CardContent>
          </Card>
        ))}

        {skills.length === 0 && (
          <div className="col-span-full text-center py-12 text-muted-foreground">
            No skills found. Create one to get started!
          </div>
        )}
      </div>
    </div>
  );
}
