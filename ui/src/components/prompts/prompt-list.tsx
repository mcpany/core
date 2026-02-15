import { useState } from "react";
import { PromptDefinition } from "@/lib/client";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Plus, Search, MessageSquare, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface PromptListProps {
    prompts: PromptDefinition[];
    selectedPrompt: PromptDefinition | null;
    onSelect: (prompt: PromptDefinition) => void;
    onCreate: () => void;
}

export function PromptList({ prompts, selectedPrompt, onSelect, onCreate }: PromptListProps) {
    const [searchQuery, setSearchQuery] = useState("");

    const filteredPrompts = prompts.filter(p =>
        p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        p.description?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex flex-col h-full bg-muted/10 border-r w-[300px] md:w-[350px] shrink-0">
            <div className="p-4 border-b space-y-3">
                <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-sm flex items-center gap-2">
                        <MessageSquare className="h-4 w-4" /> Prompt Library
                    </h3>
                    <Button variant="ghost" size="icon" onClick={onCreate} title="New Prompt">
                        <Plus className="h-4 w-4" />
                    </Button>
                </div>
                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search prompts..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8 h-9 text-sm"
                    />
                </div>
            </div>
            <ScrollArea className="flex-1">
                <div className="flex flex-col p-2 gap-1">
                    {filteredPrompts.map((prompt) => (
                        <button
                            key={prompt.name}
                            onClick={() => onSelect(prompt)}
                            className={cn(
                                "flex flex-col items-start gap-1 p-3 rounded-md text-left transition-colors hover:bg-accent hover:text-accent-foreground",
                                selectedPrompt?.name === prompt.name ? "bg-accent text-accent-foreground shadow-sm" : ""
                            )}
                        >
                            <div className="flex items-center justify-between w-full">
                                <span className="font-medium text-sm truncate">{prompt.name}</span>
                                {selectedPrompt?.name === prompt.name && <ChevronRight className="h-3 w-3 opacity-50" />}
                            </div>
                            {prompt.description && (
                                <p className="text-xs text-muted-foreground line-clamp-2">
                                    {prompt.description}
                                </p>
                            )}
                        </button>
                    ))}
                    {filteredPrompts.length === 0 && (
                        <div className="p-8 text-center text-sm text-muted-foreground">
                            {prompts.length === 0 ? "No prompts found." : "No matching prompts."}
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
