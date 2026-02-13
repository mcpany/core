// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Skill, SkillService } from '@/lib/skill-service';
import { toast } from 'sonner';
import { ChevronRight, ChevronLeft, Save, Loader2 } from 'lucide-react';
import { apiClient, ToolDefinition } from '@/lib/client';
import { MultiSelect, Option } from '@/components/ui/multi-select';
import { MarkdownEditor } from '@/components/markdown-editor';

const STEPS = ['Metadata', 'Instructions', 'Assets'];

/**
 * SkillWizard component.
 * @returns The rendered component.
 */
export default function SkillWizard() {
  const params = useParams();
  const name = params?.name as string | undefined;
  const router = useRouter();
  const isEdit = !!name;

  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [skill, setSkill] = useState<Skill>({
    name: '',
    description: '',
    instructions: '# Skill Instructions\n\nProvide step-by-step instructions for the model here.',
    allowedTools: [],
    assets: [],
  });
  const [files, setFiles] = useState<File[]>([]);
  const [availableTools, setAvailableTools] = useState<Option[]>([]);
  const [toolsLoading, setToolsLoading] = useState(false);

  useEffect(() => {
    if (isEdit && name) {
      loadSkill(name);
    }
    loadTools();
  }, [name]);

  const loadSkill = async (skillName: string) => {
    try {
      setLoading(true);
      const data = await SkillService.get(skillName);
      setSkill(data);
    } catch (err: any) {
      toast.error('Failed to load skill: ' + err.message);
      router.push('/skills');
    } finally {
      setLoading(false);
    }
  };

  const loadTools = async () => {
      setToolsLoading(true);
      try {
          const data = await apiClient.listTools();
          const options: Option[] = (data.tools || []).map((t: ToolDefinition) => ({
              label: t.name,
              value: t.name,
              description: t.description ? (t.description.length > 50 ? t.description.substring(0, 50) + "..." : t.description) : undefined
          }));
          setAvailableTools(options);
      } catch (e) {
          console.error("Failed to load tools", e);
          toast.error("Failed to load available tools");
      } finally {
          setToolsLoading(false);
      }
  };

  const handleChange = (field: keyof Skill, value: any) => {
    setSkill((prev) => ({ ...prev, [field]: value }));
  };

  const handleNext = () => {
    if (currentStep < STEPS.length - 1) {
      setCurrentStep((prev) => prev + 1);
    }
  };

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep((prev) => prev - 1);
    }
  };

  const handleSave = async () => {
    try {
      setLoading(true);
      if (isEdit && name) {
        await SkillService.update(name, skill);
        toast.success('Skill updated successfully');
      } else {
        await SkillService.create(skill);
        toast.success('Skill created successfully');
      }

      // Upload pending files if any
      if (files.length > 0) {
        const targetName = isEdit ? name : skill.name;
        for (const file of files) {
           await SkillService.uploadAsset(targetName!, `scripts/${file.name}`, file);
        }
        toast.success('Assets uploaded');
      }

      router.push('/skills');
    } catch (err: any) {
      toast.error('Failed to save skill: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles(Array.from(e.target.files));
    }
  };

  if (loading && isEdit && !skill.name) {
     return (
         <div className="flex items-center justify-center h-64">
             <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
         </div>
     );
  }

  return (
    <div className="container mx-auto py-8 max-w-3xl">
      <Card>
        <CardHeader>
          <CardTitle>{isEdit ? 'Edit Skill' : 'Create New Skill'}</CardTitle>
          <div className="flex gap-2 mt-4">
            {STEPS.map((step, idx) => (
              <div
                key={step}
                className={`flex-1 h-2 rounded-full transition-colors ${
                  idx <= currentStep ? 'bg-primary' : 'bg-secondary'
                }`}
              />
            ))}
          </div>
          <div className="text-center text-sm text-muted-foreground mt-2">
            Step {currentStep + 1}: {STEPS[currentStep]}
          </div>
        </CardHeader>
        <CardContent className="py-4">
          {currentStep === 0 && (
            <div className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Skill Name (ID)</Label>
                <Input
                  id="name"
                  value={skill.name}
                  onChange={(e) => handleChange('name', e.target.value)}
                  placeholder="e.g. data-processing"
                  disabled={isEdit || loading}
                />
                <p className="text-xs text-muted-foreground">
                  Lowercase alphanumeric and hyphens only.
                </p>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={skill.description}
                  onChange={(e) => handleChange('description', e.target.value)}
                  placeholder="Briefly describe what this skill does."
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="tools">Allowed Tools</Label>
                {toolsLoading ? (
                    <div className="text-sm text-muted-foreground flex items-center gap-2">
                        <Loader2 className="h-3 w-3 animate-spin" /> Loading tools...
                    </div>
                ) : (
                    <MultiSelect
                        options={availableTools}
                        selected={skill.allowedTools || []}
                        onChange={(selected) => handleChange('allowedTools', selected)}
                        placeholder="Select tools..."
                    />
                )}
                <p className="text-xs text-muted-foreground">
                    Select the tools that this skill is permitted to use.
                </p>
              </div>
            </div>
          )}

          {currentStep === 1 && (
            <div className="space-y-4 h-[500px] flex flex-col">
              <Label>Instructions</Label>
              <MarkdownEditor
                className="flex-1"
                value={skill.instructions}
                onChange={(val) => handleChange('instructions', val)}
                placeholder="# Skill Instructions..."
              />
            </div>
          )}

          {currentStep === 2 && (
            <div className="space-y-4">
               <div>
                  <h3 className="font-medium mb-2">Existing Assets</h3>
                  {skill.assets && skill.assets.length > 0 ? (
                      <ul className="list-disc pl-5">
                          {skill.assets.map(a => <li key={a} className="text-sm font-mono">{a}</li>)}
                      </ul>
                  ) : <p className="text-sm text-muted-foreground italic">No assets uploaded.</p>}
               </div>

               <div className="border-t pt-4">
                  <Label htmlFor="file-upload">Upload New Assets</Label>
                  <div className="flex gap-2 items-center mt-2">
                     <Input id="file-upload" type="file" multiple onChange={handleFileSelect} />
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                      Scripts, templates, or data files required by the skill.
                  </p>
               </div>
            </div>
          )}
        </CardContent>
        <CardFooter className="flex justify-between">
          <Button variant="outline" onClick={handlePrev} disabled={currentStep === 0}>
            <ChevronLeft className="mr-2 h-4 w-4" /> Back
          </Button>

          {currentStep < STEPS.length - 1 ? (
            <Button onClick={handleNext}>
              Next <ChevronRight className="ml-2 h-4 w-4" />
            </Button>
          ) : (
            <Button onClick={handleSave} disabled={loading}>
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
              {isEdit ? 'Update Skill' : 'Create Skill'}
            </Button>
          )}
        </CardFooter>
      </Card>
    </div>
  );
}
