// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  CardContent,
} from "@/components/ui/card";
import {
  Plus,
  Trash2,
  Edit,
  Loader2,
  Layers,
  Wrench,
  FileArchive,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Skill, SkillService } from "@/lib/skill-service";
import { Link } from "react-router-dom";
import { toast } from "sonner";

/**
 * SkillList component.
 * @returns The rendered component.
 */
export default function SkillList() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(true);
  const [skillToDelete, setSkillToDelete] = useState<string | null>(null);

  useEffect(() => {
    loadSkills();
  }, []);

  const loadSkills = async () => {
    try {
      const list = await SkillService.list();
      setSkills(list);
    } catch (err: any) {
      toast.error("Failed to load skills: " + err.message);
    } finally {
      setLoading(false);
    }
  };

  const confirmDelete = async () => {
    if (!skillToDelete) return;
    try {
      await SkillService.delete(skillToDelete);
      toast.success("Skill deleted");
      loadSkills();
    } catch (err: any) {
      toast.error("Failed to delete skill: " + err.message);
    } finally {
      setSkillToDelete(null);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="animate-spin h-8 w-8 text-muted-foreground" />
      </div>
    );
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

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {skills.map((skill) => (
          <Card
            key={skill.name}
            className="backdrop-blur-sm bg-background/50 border-muted-foreground/20 hover:border-primary/50 transition-all duration-300"
          >
            <CardHeader>
              <CardTitle className="flex justify-between items-start gap-4">
                <span className="break-all">{skill.name}</span>
                <div className="flex gap-2">
                  <Link to={`/skills/${skill.name}/edit`}>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <Edit className="h-4 w-4" />
                    </Button>
                  </Link>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                    onClick={() => setSkillToDelete(skill.name)}
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
              <div className="flex flex-wrap items-center gap-2 pt-2">
                {skill.allowedTools && skill.allowedTools.length > 0 && (
                  <Badge
                    variant="secondary"
                    className="font-normal text-xs px-2 py-0.5"
                  >
                    <Wrench className="h-3 w-3 mr-1" />
                    {skill.allowedTools.length} Tools
                  </Badge>
                )}
                {skill.assets && skill.assets.length > 0 && (
                  <Badge
                    variant="outline"
                    className="font-normal text-xs px-2 py-0.5 text-muted-foreground"
                  >
                    <FileArchive className="h-3 w-3 mr-1" />
                    {skill.assets.length} Assets
                  </Badge>
                )}
                <Link
                  to={`/skills/${skill.name}`}
                  className="ml-auto text-xs text-primary hover:underline font-medium"
                >
                  View Details
                </Link>
              </div>
            </CardContent>
          </Card>
        ))}

        {skills.length === 0 && (
          <div className="col-span-full flex flex-col items-center justify-center py-16 text-center">
            <Layers className="h-12 w-12 text-muted-foreground mb-4 opacity-50" />
            <h3 className="text-lg font-medium">No skills found</h3>
            <p className="text-sm text-muted-foreground mb-6 max-w-sm">
              Skills are reusable capabilities combining tools and prompts.
              Create one to get started!
            </p>
            <Link to="/skills/create">
              <Button>
                <Plus className="mr-2 h-4 w-4" /> Create Skill
              </Button>
            </Link>
          </div>
        )}
      </div>

      <AlertDialog
        open={!!skillToDelete}
        onOpenChange={(open) => !open && setSkillToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Skill</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the skill "{skillToDelete}"? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
