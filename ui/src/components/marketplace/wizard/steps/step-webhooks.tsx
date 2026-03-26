/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useWizard } from "../wizard-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Plus, Trash2, Webhook } from "lucide-react";
<<<<<<< HEAD
import { WizardWebhook } from "../wizard-context";
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

/**
 * StepWebhooks component.
 * @returns The rendered component.
 */
export function StepWebhooks() {
  const { state, updateConfig } = useWizard();
  const { config } = state;

  const addWebhook = (type: 'preCallHooks' | 'postCallHooks') => {
<<<<<<< HEAD
      const newHook: WizardWebhook = {
=======
      const newHook = {
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
          name: `webhook-${Date.now()}`,
          webhook: {
              url: "https://",
              timeout: "5s",
              webhookSecret: ""
          }
<<<<<<< HEAD
      };

      const hooks = (config[type] as WizardWebhook[]) ? [...(config[type] as WizardWebhook[])] : [];
=======
      } as any;

      const hooks = config[type] ? [...config[type]] : [];
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
      hooks.push(newHook);

      updateConfig({
          [type]: hooks
      });
  };

  const removeWebhook = (type: 'preCallHooks' | 'postCallHooks', index: number) => {
<<<<<<< HEAD
      const hooks = (config[type] as WizardWebhook[]) ? [...(config[type] as WizardWebhook[])] : [];
=======
      const hooks = config[type] ? [...config[type]] : [];
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
      hooks.splice(index, 1);
      updateConfig({
          [type]: hooks
      });
  };

  const updateWebhook = (type: 'preCallHooks' | 'postCallHooks', index: number, field: string, value: string) => {
<<<<<<< HEAD
      const hooks = (config[type] as WizardWebhook[]) ? [...(config[type] as WizardWebhook[])] : [];
      const hook = { ...hooks[index] };
=======
      const hooks = config[type] ? [...config[type]] : [];
      const hook = { ...(hooks[index] as unknown as Record<string, unknown>) };
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

      if (field === 'name') {
          hook.name = value;
      } else {
<<<<<<< HEAD
           const currentWebhook = hook.webhook || { url: "", webhookSecret: "", timeout: "" };
           hook.webhook = { ...currentWebhook, [field]: value };
      }

      hooks[index] = hook;
=======
           const currentWebhook = hook.webhook || { url: "", webhookSecret: "" };
           hook.webhook = { ...currentWebhook, [field]: value } as any;
      }

      hooks[index] = hook as any;
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
      updateConfig({
          [type]: hooks
      });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2">
          <Label className="text-lg font-semibold flex items-center gap-2">
            <Webhook className="h-5 w-5" />
            Pre-Call Webhooks
          </Label>
          <p className="text-sm text-muted-foreground">
              Webhooks executed before the upstream service call. Can be used for policy checks or argument transformation.
          </p>

          <div className="space-y-4">
<<<<<<< HEAD
              {(config.preCallHooks as WizardWebhook[])?.map((hook: WizardWebhook, idx: number) => (
=======
              {config.preCallHooks?.map((hook: any, idx: number) => (
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
                  <Card key={idx}>
                      <CardContent className="pt-6 grid gap-4">
                          <div className="flex justify-between items-start">
                              <div className="grid gap-2 flex-1 mr-4">
                                  <Label>Hook Name</Label>
                                  <Input
                                    value={hook.name}
                                    onChange={(e) => updateWebhook('preCallHooks', idx, 'name', e.target.value)}
                                  />
                              </div>
                              <Button variant="ghost" size="icon" onClick={() => removeWebhook('preCallHooks', idx)}>
                                  <Trash2 className="h-4 w-4 text-destructive" />
                              </Button>
                          </div>
                          <div className="grid grid-cols-2 gap-4">
                              <div className="space-y-2">
                                  <Label>URL</Label>
                                  <Input
                                    value={hook.webhook?.url || ''}
                                    onChange={(e) => updateWebhook('preCallHooks', idx, 'url', e.target.value)}
                                    placeholder="https://api.example.com/webhook"
                                  />
                              </div>
                              <div className="space-y-2">
                                  <Label>Timeout</Label>
                                  <Input
                                    value={hook.webhook?.timeout || ''}
                                    onChange={(e) => updateWebhook('preCallHooks', idx, 'timeout', e.target.value)}
                                    placeholder="5s"
                                  />
                              </div>
                          </div>
                          <div className="space-y-2">
                              <Label>Secret (Optional)</Label>
                               <Input
                                    type="password"
                                    value={hook.webhook?.webhookSecret || ''}
                                    onChange={(e) => updateWebhook('preCallHooks', idx, 'webhookSecret', e.target.value)}
                                    placeholder="Signing secret"
                                  />
                          </div>
                      </CardContent>
                  </Card>
              ))}
              <Button variant="outline" onClick={() => addWebhook('preCallHooks')} className="w-full border-dashed">
                  <Plus className="mr-2 h-4 w-4" /> Add Pre-Call Webhook
              </Button>
          </div>
      </div>

      <div className="space-y-2">
          <Label className="text-lg font-semibold flex items-center gap-2">
             <Webhook className="h-5 w-5" />
             Post-Call Webhooks
          </Label>
           <p className="text-sm text-muted-foreground">
              Webhooks executed after the upstream service call. Can be used for logging or result transformation.
          </p>
           <div className="space-y-4">
<<<<<<< HEAD
              {(config.postCallHooks as WizardWebhook[])?.map((hook: WizardWebhook, idx: number) => (
=======
              {config.postCallHooks?.map((hook: any, idx: number) => (
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
                  <Card key={idx}>
                      <CardContent className="pt-6 grid gap-4">
                          <div className="flex justify-between items-start">
                              <div className="grid gap-2 flex-1 mr-4">
                                  <Label>Hook Name</Label>
                                  <Input
                                    value={hook.name}
                                    onChange={(e) => updateWebhook('postCallHooks', idx, 'name', e.target.value)}
                                  />
                              </div>
                              <Button variant="ghost" size="icon" onClick={() => removeWebhook('postCallHooks', idx)}>
                                  <Trash2 className="h-4 w-4 text-destructive" />
                              </Button>
                          </div>
                          <div className="grid grid-cols-2 gap-4">
                              <div className="space-y-2">
                                  <Label>URL</Label>
                                  <Input
                                    value={hook.webhook?.url || ''}
                                    onChange={(e) => updateWebhook('postCallHooks', idx, 'url', e.target.value)}
                                    placeholder="https://api.example.com/webhook"
                                  />
                              </div>
                              <div className="space-y-2">
                                  <Label>Timeout</Label>
                                  <Input
                                    value={hook.webhook?.timeout || ''}
                                    onChange={(e) => updateWebhook('postCallHooks', idx, 'timeout', e.target.value)}
                                    placeholder="5s"
                                  />
                              </div>
                          </div>
                           <div className="space-y-2">
                              <Label>Secret (Optional)</Label>
                               <Input
                                    type="password"
                                    value={hook.webhook?.webhookSecret || ''}
                                    onChange={(e) => updateWebhook('postCallHooks', idx, 'webhookSecret', e.target.value)}
                                    placeholder="Signing secret"
                                  />
                          </div>
                      </CardContent>
                  </Card>
              ))}
              <Button variant="outline" onClick={() => addWebhook('postCallHooks')} className="w-full border-dashed">
                  <Plus className="mr-2 h-4 w-4" /> Add Post-Call Webhook
              </Button>
          </div>
      </div>

      <div className="pt-6 border-t">
          <Label className="text-lg font-semibold flex items-center gap-2 mb-4">
             <div className="bg-primary/10 p-2 rounded-full"><Webhook className="h-4 w-4" /></div>
             Transformers Playground
          </Label>
          <Card className="bg-muted/30">
              <CardContent className="p-4 space-y-4">
                  <p className="text-sm text-muted-foreground">
                      Simulate how your webhooks and transformers will modify the request/response.
                  </p>
                  <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                          <Label>Sample Input (JSON)</Label>
                          <textarea
                            className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                            placeholder='{"message": "hello"}'
                            defaultValue='{"message": "hello", "user": "test"}'
                            id="playground-input"
                          />
                      </div>
                      <div className="space-y-2">
                          <Label>Simulated Output</Label>
                          <div id="playground-output" className="min-h-[120px] w-full rounded-md border border-input bg-muted/50 px-3 py-2 text-sm font-mono text-muted-foreground">
                              Click Test to see result...
                          </div>
                      </div>
                  </div>
                  <Button onClick={() => {
                      const inputEl = document.getElementById('playground-input') as HTMLTextAreaElement;
                      const outputEl = document.getElementById('playground-output');
                      try {
                          const val = JSON.parse(inputEl.value);
                          // Mock transformation: add timestamp and processed flag
                          val._processed_at = new Date().toISOString();
                          val._transformed = true;
                          if (config.preCallHooks?.length) {
<<<<<<< HEAD
                              val._hooks_applied = (config.preCallHooks as WizardWebhook[]).map((h: WizardWebhook) => h.name);
=======
                              val._hooks_applied = config.preCallHooks.map((h: any) => h.name);
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
                          }
                          if (outputEl) outputEl.textContent = JSON.stringify(val, null, 2);
                      } catch (e) {
                           if (outputEl) outputEl.textContent = "Error: Invalid JSON input";
                      }
                  }}>
                      Test Transformation
                  </Button>
              </CardContent>
          </Card>
      </div>
    </div>
  );
}
