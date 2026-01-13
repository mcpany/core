import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SERVICE_TEMPLATES, ServiceTemplate } from "@/data/service-templates";
import { Database, Globe, Github, MessageSquare, HardDrive, LayoutTemplate, LucideIcon } from "lucide-react";

interface TemplateSelectorProps {
    onSelect: (template: ServiceTemplate) => void;
    onCancel: () => void;
}

const IconMap: Record<string, LucideIcon> = {
    "Globe": Globe,
    "Database": Database,
    "Github": Github,
    "MessageSquare": MessageSquare,
    "FileDatabase": HardDrive,
};

export function TemplateSelector({ onSelect, onCancel }: TemplateSelectorProps) {
    return (
        <div className="space-y-4">
             <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Empty Template */}
                <Card
                    className="cursor-pointer hover:border-primary transition-colors flex flex-col justify-between"
                    onClick={() => onSelect({
                        id: "custom",
                        name: "Custom Service",
                        description: "Configure a service from scratch.",
                        icon: "LayoutTemplate",
                        tags: ["Custom"],
                        config: {
                             id: "",
                             name: "",
                             version: "1.0.0",
                             disable: false,
                             priority: 0,
                             loadBalancingStrategy: 0,
                             httpService: { address: "" }
                        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        } as any
                    })}
                >
                    <CardHeader className="flex flex-row items-start gap-4 pb-2">
                         <div className="p-2 bg-primary/10 rounded-lg">
                            <LayoutTemplate className="h-6 w-6 text-primary" />
                         </div>
                         <div className="space-y-1">
                            <CardTitle className="text-base">Custom Service</CardTitle>
                            <CardDescription className="line-clamp-2">
                                Start with a blank configuration.
                            </CardDescription>
                         </div>
                    </CardHeader>
                </Card>

                {SERVICE_TEMPLATES.map((template) => {
                    const Icon = IconMap[template.icon] || Globe;
                    return (
                        <Card
                            key={template.id}
                            className="cursor-pointer hover:border-primary transition-colors flex flex-col justify-between"
                            onClick={() => onSelect(template)}
                        >
                            <CardHeader className="flex flex-row items-start gap-4 pb-2">
                                <div className="p-2 bg-muted rounded-lg">
                                    <Icon className="h-6 w-6 text-muted-foreground" />
                                </div>
                                <div className="space-y-1">
                                    <CardTitle className="text-base">{template.name}</CardTitle>
                                    <CardDescription className="line-clamp-2">
                                        {template.description}
                                    </CardDescription>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="flex flex-wrap gap-2">
                                    {template.tags.map(tag => (
                                        <Badge key={tag} variant="secondary" className="text-xs">
                                            {tag}
                                        </Badge>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    );
                })}
             </div>
             <div className="flex justify-end pt-4">
                 <Button variant="ghost" onClick={onCancel}>Cancel</Button>
             </div>
        </div>
    );
}
