# UI Development Guidelines

## Markdown Rendering

When rendering user-generated content or assistant responses, avoid using raw `<p>` tags. Instead, use the `MarkdownRenderer` component to ensure proper formatting of code blocks, tables, and rich text.

```tsx
import { MarkdownRenderer } from "@/components/ui/markdown-renderer";

<MarkdownRenderer content={message.content} />
```

This component handles:
- GitHub Flavored Markdown (tables, strikethrough)
- Syntax highlighting for code blocks (using `react-syntax-highlighter`)
- Tailwind typography styling (prose)
