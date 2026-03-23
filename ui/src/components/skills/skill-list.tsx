// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0



import React, { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardDescription, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Plus, Trash2, Edit } from 'lucide-react';
import { Skill, SkillService } from '@/lib/skill-service';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { Checkbox } from '@/components/ui/checkbox';

/**
 * SkillList component.
 * @returns The rendered component.
 */
export default function SkillList() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadSkills();
  }, []);

  const loadSkills = async () => {
    try {
      const list = await SkillService.list();
      setSkills(list);
    } catch (err: unknown) {
      toast.error('Failed to load skills: ' + (err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Are you sure you want to delete the skill "${name}"?`)) return;
    try {
      await SkillService.delete(name);
      toast.success('Skill deleted');
      setSelectedSkills(prev => {
        const next = new Set(prev);
        next.delete(name);
        return next;
      });
      loadSkills();
    } catch (err: unknown) {
      toast.error('Failed to delete skill: ' + (err as Error).message);
    }
  };

  const handleBulkDelete = async () => {
    if (selectedSkills.size === 0) return;
    if (!confirm(`Are you sure you want to delete ${selectedSkills.size} selected skills?`)) return;

    let deletedCount = 0;
    for (const name of selectedSkills) {
      try {
        await SkillService.delete(name);
        deletedCount++;
      } catch (err: unknown) {
        toast.error(`Failed to delete skill "${name}": ` + (err as Error).message);
      }
    }

    if (deletedCount > 0) {
      toast.success(`Successfully deleted ${deletedCount} skill(s)`);
      setSelectedSkills(new Set());
      loadSkills();
    }
  };

  const toggleSelection = (name: string) => {
    setSelectedSkills(prev => {
      const next = new Set(prev);
      if (next.has(name)) {
        next.delete(name);
      } else {
        next.add(name);
      }
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedSkills.size === skills.length) {
      setSelectedSkills(new Set());
    } else {
      setSelectedSkills(new Set(skills.map(s => s.name)));
    }
  };

  if (loading) {
    return <div className="p-4">Loading skills...</div>;
  }

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Agent Skills</h1>
        <div className="flex items-center gap-2">
          {selectedSkills.size > 0 && (
            <div className="flex items-center gap-2 mr-4 p-1 bg-muted rounded-md px-3 border fade-in animate-in">
              <span className="text-sm font-medium">{selectedSkills.size} selected</span>
              <Button variant="destructive" size="sm" onClick={handleBulkDelete} className="h-7 text-xs px-2">
                <Trash2 className="mr-1.5 h-3 w-3" /> Bulk Delete
              </Button>
            </div>
          )}
          {skills.length > 0 && (
            <Button variant="outline" size="sm" onClick={toggleSelectAll} className="mr-2 h-9">
              {selectedSkills.size === skills.length ? 'Deselect All' : 'Select All'}
            </Button>
          )}
          <Link to="/skills/create">
            <Button>
              <Plus className="mr-2 h-4 w-4" /> Create Skill
            </Button>
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {skills.map((skill) => (
          <Card key={skill.name} className={`hover:shadow-lg transition-all ${selectedSkills.has(skill.name) ? 'border-primary ring-1 ring-primary/20 bg-primary/5' : ''}`}>
            <CardHeader>
              <CardTitle className="flex justify-between items-center">
                <div className="flex items-center gap-3">
                  <Checkbox
                    checked={selectedSkills.has(skill.name)}
                    onCheckedChange={() => toggleSelection(skill.name)}
                    aria-label={`Select skill ${skill.name}`}
                  />
                  <span>{skill.name}</span>
                </div>
                <div className="flex gap-2">
                  <Link to={`/skills/${skill.name}/edit`}>
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
                <Link to={`/skills/${skill.name}`} className="ml-auto text-primary hover:underline">
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
