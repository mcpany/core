/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, ProfileDefinition } from "@/lib/client";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { EnvVarEditor } from "./env-var-editor";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";

interface ProfileDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  profile: ProfileDefinition | null;
  onSave: () => void;
}

export function ProfileDialog({ open, onOpenChange, profile, onSave }: ProfileDialogProps) {
  const [name, setName] = useState("");
  const [requiredRoles, setRequiredRoles] = useState<string[]>([]);
  const [parentProfileIds, setParentProfileIds] = useState<string[]>([]);
  const [secrets, setSecrets] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [newRole, setNewRole] = useState("");
  const [newParent, setNewParent] = useState("");

  const { toast } = useToast();

  useEffect(() => {
    if (profile) {
      setName(profile.name);
      setRequiredRoles(profile.requiredRoles || []);
      setParentProfileIds(profile.parentProfileIds || []);

      // Convert secrets map (key -> SecretValue) to key -> value string
      const secretsMap: Record<string, string> = {};
      if (profile.secrets) {
        Object.entries(profile.secrets).forEach(([k, v]) => {
          secretsMap[k] = v.plainText || v.environmentVariable || v.filePath || "";
        });
      }
      setSecrets(secretsMap);
    } else {
      setName("");
      setRequiredRoles([]);
      setParentProfileIds([]);
      setSecrets({});
    }
  }, [profile, open]);

  const handleSubmit = async () => {
    if (!name) {
      toast({ title: "Error", description: "Profile name is required", variant: "destructive" });
      return;
    }

    setIsSubmitting(true);
    try {
      // Convert flat secrets back to SecretValue objects
      const secretsObj: Record<string, any> = {};
      Object.entries(secrets).forEach(([k, v]) => {
        secretsObj[k] = { plainText: v };
      });

      const payload: ProfileDefinition = {
        name,
        requiredRoles,
        parentProfileIds,
        secrets: secretsObj,
        // Preserve other fields if needed, but for now simplistic
        selector: undefined,
        serviceConfig: undefined
      };

      if (profile) {
        await apiClient.updateProfile(payload);
        toast({ title: "Success", description: "Profile updated successfully" });
      } else {
        await apiClient.createProfile(payload);
        toast({ title: "Success", description: "Profile created successfully" });
      }
      onSave();
    } catch (e: any) {
      console.error(e);
      toast({ title: "Error", description: e.message || "Failed to save profile", variant: "destructive" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const addRole = () => {
    if (newRole && !requiredRoles.includes(newRole)) {
      setRequiredRoles([...requiredRoles, newRole]);
      setNewRole("");
    }
  };

  const removeRole = (role: string) => {
    setRequiredRoles(requiredRoles.filter(r => r !== role));
  };

  const addParent = () => {
    if (newParent && !parentProfileIds.includes(newParent)) {
      setParentProfileIds([...parentProfileIds, newParent]);
      setNewParent("");
    }
  };

  const removeParent = (pid: string) => {
    setParentProfileIds(parentProfileIds.filter(p => p !== pid));
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{profile ? "Edit Profile" : "Create Profile"}</DialogTitle>
          <DialogDescription>
            Configure execution context, roles, and environment variables.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Profile Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. production"
              disabled={!!profile} // ID/Name usually immutable for Update if used as key
            />
            {profile && <p className="text-xs text-muted-foreground">Profile name cannot be changed.</p>}
          </div>

          <Tabs defaultValue="secrets" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="secrets">Secrets & Env</TabsTrigger>
              <TabsTrigger value="roles">Access Control</TabsTrigger>
              <TabsTrigger value="parents">Inheritance</TabsTrigger>
            </TabsList>

            <TabsContent value="secrets" className="space-y-4 pt-4">
              <div className="space-y-2">
                <Label>Environment Variables</Label>
                <p className="text-xs text-muted-foreground mb-2">
                  These variables will be injected into the execution context for tools running under this profile.
                </p>
                <EnvVarEditor value={secrets} onChange={setSecrets} />
              </div>
            </TabsContent>

            <TabsContent value="roles" className="space-y-4 pt-4">
              <div className="space-y-2">
                <Label>Required Roles</Label>
                <div className="flex gap-2">
                  <Input
                    value={newRole}
                    onChange={(e) => setNewRole(e.target.value)}
                    placeholder="Add role..."
                    onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), addRole())}
                  />
                  <Button onClick={addRole} variant="secondary">Add</Button>
                </div>
                <div className="flex flex-wrap gap-2 mt-2">
                  {requiredRoles.map(role => (
                    <Badge key={role} variant="secondary" className="gap-1 pr-1">
                      {role}
                      <button onClick={() => removeRole(role)} className="hover:bg-muted rounded-full p-0.5">
                        <X className="h-3 w-3" />
                      </button>
                    </Badge>
                  ))}
                  {requiredRoles.length === 0 && <span className="text-xs text-muted-foreground italic">No specific roles required (public to authenticated users).</span>}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="parents" className="space-y-4 pt-4">
              <div className="space-y-2">
                <Label>Parent Profiles</Label>
                <div className="flex gap-2">
                  <Input
                    value={newParent}
                    onChange={(e) => setNewParent(e.target.value)}
                    placeholder="Parent profile ID..."
                    onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), addParent())}
                  />
                  <Button onClick={addParent} variant="secondary">Add</Button>
                </div>
                <div className="flex flex-wrap gap-2 mt-2">
                  {parentProfileIds.map(pid => (
                    <Badge key={pid} variant="outline" className="gap-1 pr-1">
                      {pid}
                      <button onClick={() => removeParent(pid)} className="hover:bg-muted rounded-full p-0.5">
                        <X className="h-3 w-3" />
                      </button>
                    </Badge>
                  ))}
                  {parentProfileIds.length === 0 && <span className="text-xs text-muted-foreground italic">No parent profiles.</span>}
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={isSubmitting}>
            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Save Profile
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
