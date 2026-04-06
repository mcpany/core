// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0



import React, { useEffect, useState, useMemo, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardDescription, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Plus, Trash2, Edit, Search, LayoutGrid, List } from 'lucide-react';
import { Skill, SkillService } from '@/lib/skill-service';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';

/**
 * Intent: Document SkillList
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * SkillList component.
 * @returns The rendered component.
 */
export default function SkillList() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [selected, setSelected] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadSkills();
  }, []);

  const loadSkills = async () => {
    try {
      const list = await SkillService.list();
      setSkills(list);
      setSelected(new Set()); // Reset selection when list updates
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

  const handleBulkDelete = async () => {
    if (!confirm(`Are you sure you want to delete ${selected.size} skills?`)) return;
    try {
      await Promise.all(Array.from(selected).map(name => SkillService.delete(name)));
      toast.success(`${selected.size} skills deleted`);
      loadSkills();
    } catch (err: any) {
      toast.error('Failed to delete some skills: ' + err.message);
      loadSkills(); // Reload to get actual state
    }
  };

  const filteredSkills = useMemo(() => {
    if (!searchQuery) return skills;
    const lowerQuery = searchQuery.toLowerCase();
    return skills.filter(
      (skill) =>
        skill.name.toLowerCase().includes(lowerQuery) ||
        skill.description?.toLowerCase().includes(lowerQuery)
    );
  }, [skills, searchQuery]);

  // Handle Select All based on filtered items
  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelected(new Set(filteredSkills.map(s => s.name)));
    } else {
      setSelected(new Set());
    }
  }, [filteredSkills]);

  const handleSelectOne = useCallback((name: string, checked: boolean) => {
    setSelected(prev => {
        const newSelected = new Set(prev);
        if (checked) {
          newSelected.add(name);
        } else {
          newSelected.delete(name);
        }
        return newSelected;
    });
  }, []);

  const isAllSelected = filteredSkills.length > 0 && selected.size === filteredSkills.length;

  if (loading) {
    return <div className="p-4">Loading skills...</div>;
  }

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Agent Skills</h1>
        <Link to="/skills/create">
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Create Skill
          </Button>
        </Link>
      </div>

      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-6 relative">
        <div className="relative w-full md:w-[300px]">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search skills..."
            className="pl-8"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className="flex items-center gap-2">
          {selected.size > 0 && (
            <div className="flex items-center gap-2 mr-4 bg-muted/40 px-3 py-1.5 rounded-md border shadow-sm animate-in fade-in slide-in-from-right-4">
              <span className="text-sm font-medium">{selected.size} selected</span>
              <div className="h-4 w-px bg-border mx-1"></div>
              <Button size="sm" variant="destructive" onClick={handleBulkDelete}>
                <Trash2 className="mr-2 h-4 w-4" /> Bulk Delete
              </Button>
            </div>
          )}
          <div className="flex items-center bg-muted rounded-md p-1 border">
            <Button
              variant={viewMode === "grid" ? "secondary" : "ghost"}
              size="icon"
              className="h-8 w-8"
              onClick={() => setViewMode("grid")}
            >
              <LayoutGrid className="h-4 w-4" />
            </Button>
            <Button
              variant={viewMode === "list" ? "secondary" : "ghost"}
              size="icon"
              className="h-8 w-8"
              onClick={() => setViewMode("list")}
            >
              <List className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {viewMode === "list" ? (
        <div className="rounded-md border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[50px] text-center">
                  <Checkbox
                    checked={isAllSelected}
                    onCheckedChange={handleSelectAll}
                    aria-label="Select all skills"
                  />
                </TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Tools</TableHead>
                <TableHead>Assets</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredSkills.length === 0 ? (
                 <TableRow>
                  <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                    No skills found matching your search.
                  </TableCell>
                </TableRow>
              ) : (
                filteredSkills.map((skill) => (
                  <TableRow key={skill.name} className={selected.has(skill.name) ? "bg-muted/50" : ""}>
                    <TableCell className="text-center">
                      <Checkbox
                        checked={selected.has(skill.name)}
                        onCheckedChange={(checked) => handleSelectOne(skill.name, !!checked)}
                        aria-label={`Select ${skill.name}`}
                      />
                    </TableCell>
                    <TableCell className="font-medium">
                      <Link to={`/skills/${skill.name}`} className="hover:underline">
                        {skill.name}
                      </Link>
                    </TableCell>
                    <TableCell className="max-w-[300px] truncate" title={skill.description}>
                      {skill.description}
                    </TableCell>
                    <TableCell>
                      {skill.allowedTools?.length || 0}
                    </TableCell>
                    <TableCell>
                      {skill.assets?.length || 0}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Link to={`/skills/${skill.name}/edit`}>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Edit className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                          onClick={() => handleDelete(skill.name)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredSkills.map((skill) => (
            <Card key={skill.name} className={`hover:shadow-lg transition-all ${selected.has(skill.name) ? 'ring-2 ring-primary border-transparent' : ''}`}>
              <CardHeader className="pb-3">
                <CardTitle className="flex justify-between items-start gap-4">
                  <div className="flex items-center gap-3 overflow-hidden">
                    <Checkbox
                      checked={selected.has(skill.name)}
                      onCheckedChange={(checked) => handleSelectOne(skill.name, !!checked)}
                      aria-label={`Select ${skill.name}`}
                      className="mt-1"
                    />
                    <Link to={`/skills/${skill.name}`} className="hover:underline truncate text-lg">
                      {skill.name}
                    </Link>
                  </div>
                  <div className="flex gap-1 shrink-0">
                    <Link to={`/skills/${skill.name}/edit`}>
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <Edit className="h-4 w-4" />
                      </Button>
                    </Link>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                      onClick={() => handleDelete(skill.name)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </CardTitle>
                <CardDescription className="line-clamp-2 h-10 mt-2 text-sm">
                  {skill.description || "No description provided."}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                  {skill.allowedTools && skill.allowedTools.length > 0 && (
                    <span className="bg-secondary px-2 py-1 rounded-md font-medium">
                      {skill.allowedTools.length} Tools
                    </span>
                  )}
                  {skill.assets && skill.assets.length > 0 && (
                    <span className="bg-secondary px-2 py-1 rounded-md font-medium">
                      {skill.assets.length} Assets
                    </span>
                  )}
                  <Link to={`/skills/${skill.name}`} className="ml-auto text-primary hover:underline font-medium">
                    View Details
                  </Link>
                </div>
              </CardContent>
            </Card>
          ))}
          {filteredSkills.length === 0 && (
            <div className="col-span-full text-center py-12 text-muted-foreground bg-muted/20 rounded-xl border border-dashed">
              No skills found matching your search.
            </div>
          )}
        </div>
      )}
    </div>
  );
}
