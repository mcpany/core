import React, { useState } from 'react';
import { ChevronRight, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';

interface JsonNodeProps {
  keyName: string | number;
  value: any;
  isLast: boolean;
  depth: number;
  defaultExpandedLevel: number;
}

const JsonNode: React.FC<JsonNodeProps> = ({ keyName, value, isLast, depth, defaultExpandedLevel }) => {
  const [isExpanded, setIsExpanded] = useState(depth < defaultExpandedLevel);

  const isObject = value !== null && typeof value === 'object';
  const isArray = Array.isArray(value);
  const isEmpty = isObject && Object.keys(value).length === 0;

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsExpanded(!isExpanded);
  };

  const renderValue = () => {
    if (value === null) return <span className="text-muted-foreground">null</span>;
    if (typeof value === 'boolean') return <span className="text-blue-500">{value ? 'true' : 'false'}</span>;
    if (typeof value === 'number') return <span className="text-green-500">{value}</span>;
    if (typeof value === 'string') return <span className="text-amber-500">"{value}"</span>;
    return <span>{String(value)}</span>;
  };

  // Skip showing the root key "" for the top-level object/array
  const displayKey = keyName !== '' ? `"${keyName}": ` : '';

  if (isObject) {
    const keys = Object.keys(value);

    return (
      <div className="font-mono text-sm leading-relaxed relative" style={{ marginLeft: depth > 0 ? '1.5rem' : 0 }}>
        <div
          className={cn("flex items-start", !isEmpty && "cursor-pointer hover:bg-muted/50 rounded -ml-5 pl-5")}
          onClick={!isEmpty ? handleToggle : undefined}
        >
          {!isEmpty && (
            <span className="absolute -left-4 top-1 w-3 h-3 flex items-center justify-center opacity-50 hover:opacity-100 transition-opacity">
              {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
            </span>
          )}

          {keyName !== '' && <span className="text-purple-600 dark:text-purple-400 mr-1">{displayKey}</span>}
          <span className="text-foreground">{isArray ? '[' : '{'}</span>
          {!isExpanded && !isEmpty && (
            <span className="text-muted-foreground mx-2 text-xs px-1.5 py-0.5 bg-muted rounded">
              {isArray ? `${keys.length} items` : `${keys.length} keys`}
            </span>
          )}
          {(!isExpanded || isEmpty) && <span className="text-foreground">{isArray ? ']' : '}'}{!isLast && ','}</span>}
        </div>

        {isExpanded && !isEmpty && (
          <div>
            {keys.map((k, i) => (
              <JsonNode
                key={k}
                keyName={isArray ? i : k}
                value={value[k as keyof typeof value]}
                isLast={i === keys.length - 1}
                depth={depth + 1}
                defaultExpandedLevel={defaultExpandedLevel}
              />
            ))}
            <div className="flex items-start -ml-5 pl-5">
              <span className="text-foreground">{isArray ? ']' : '}'}{!isLast && ','}</span>
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="font-mono text-sm leading-relaxed flex items-start" style={{ marginLeft: depth > 0 ? '1.5rem' : 0 }}>
      {keyName !== '' && <span className="text-purple-600 dark:text-purple-400 mr-1">{displayKey}</span>}
      {renderValue()}
      {!isLast && <span className="text-foreground">,</span>}
    </div>
  );
};

export interface JsonTreeProps {
  data: any;
  className?: string;
  defaultExpandedLevel?: number;
}

export const JsonTree: React.FC<JsonTreeProps> = ({ data, className, defaultExpandedLevel = 3 }) => {
  return (
    <div className={cn("bg-background border rounded-md p-4 overflow-auto", className)}>
      <JsonNode keyName="" value={data} isLast={true} depth={0} defaultExpandedLevel={defaultExpandedLevel} />
    </div>
  );
};
