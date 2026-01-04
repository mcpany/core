
"use client";

import { useState } from "react";
import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { GripVertical, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";

// Mock middleware
const initialMiddleware = [
  { id: "mw-1", name: "Authentication", type: "auth", enabled: true },
  { id: "mw-2", name: "Rate Limiter", type: "rate_limit", enabled: true },
  { id: "mw-3", name: "Logging", type: "logging", enabled: true },
  { id: "mw-4", name: "Caching", type: "cache", enabled: false },
];

export default function MiddlewarePage() {
  const [middleware, setMiddleware] = useState(initialMiddleware);

  const onDragEnd = (result: any) => {
    if (!result.destination) return;

    const items = Array.from(middleware);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    setMiddleware(items);
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h2>
         <Button>
            <Plus className="mr-2 h-4 w-4" /> Add Middleware
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                <CardTitle>Request Processing Pipeline</CardTitle>
                <CardDescription>Drag and drop to reorder the middleware execution sequence.</CardDescription>
                </CardHeader>
                <CardContent>
                    <DragDropContext onDragEnd={onDragEnd}>
                        <Droppable droppableId="middleware-list">
                        {(provided) => (
                            <div {...provided.droppableProps} ref={provided.innerRef} className="space-y-3">
                            {middleware.map((item, index) => (
                                <Draggable key={item.id} draggableId={item.id} index={index}>
                                {(provided) => (
                                    <div
                                        ref={provided.innerRef}
                                        {...provided.draggableProps}
                                        className="flex items-center p-4 bg-card border rounded-lg shadow-sm group hover:shadow-md transition-all"
                                    >
                                        <div {...provided.dragHandleProps} className="mr-4 text-muted-foreground cursor-move">
                                            <GripVertical className="h-5 w-5" />
                                        </div>
                                        <div className="flex-1">
                                            <h4 className="font-medium">{item.name}</h4>
                                            <Badge variant="outline" className="mt-1 text-xs">{item.type}</Badge>
                                        </div>
                                        <div className="flex items-center space-x-4">
                                            <Badge variant={item.enabled ? "default" : "secondary"}>
                                                {item.enabled ? "Active" : "Skipped"}
                                            </Badge>
                                        </div>
                                    </div>
                                )}
                                </Draggable>
                            ))}
                            {provided.placeholder}
                            </div>
                        )}
                        </Droppable>
                    </DragDropContext>
                </CardContent>
            </Card>
        </div>

        <div>
            <Card>
                <CardHeader>
                    <CardTitle>Available Modules</CardTitle>
                    <CardDescription>Middleware components you can add.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                    {["PII Redaction", "Input Validation", "Output Formatting", "Retry Logic"].map((mod) => (
                        <div key={mod} className="flex items-center justify-between p-3 border rounded-md hover:bg-muted/50 cursor-pointer">
                            <span className="text-sm font-medium">{mod}</span>
                            <Plus className="h-4 w-4 text-muted-foreground" />
                        </div>
                    ))}
                </CardContent>
            </Card>
        </div>
      </div>
    </div>
  );
}
