/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Monaco } from "@monaco-editor/react";

/**
 * Defines the Dracula theme for the Monaco editor.
 * @param monaco - The monaco instance.
 */
export function defineDraculaTheme(monaco: Monaco) {
  monaco.editor.defineTheme("dracula", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "", foreground: "f8f8f2", background: "282a36" },
      { token: "invalid", foreground: "ff5555" },
      { token: "emphasis", fontStyle: "italic" },
      { token: "strong", fontStyle: "bold" },
      { token: "variable", foreground: "f8f8f2" },
      { token: "variable.predefined", foreground: "f8f8f2" },
      { token: "constant", foreground: "bd93f9" },
      { token: "comment", foreground: "6272a4" },
      { token: "number", foreground: "bd93f9" },
      { token: "number.hex", foreground: "bd93f9" },
      { token: "regexp", foreground: "ffb86c" },
      { token: "annotation", foreground: "ffb86c" },
      { token: "type", foreground: "8be9fd" },
      { token: "delimiter", foreground: "f8f8f2" },
      { token: "delimiter.parenthesis", foreground: "f8f8f2" },
      { token: "delimiter.bracket", foreground: "f8f8f2" },
      { token: "delimiter.square", foreground: "f8f8f2" },
      { token: "tag", foreground: "ff79c6" },
      { token: "tag.id.jade", foreground: "ff79c6" },
      { token: "tag.class.jade", foreground: "ff79c6" },
      { token: "meta.selector.css", foreground: "ff79c6" },
      { token: "entity.other.attribute-name.css", foreground: "50fa7b" },
      { token: "entity.name.tag.scss", foreground: "ff79c6" },
      { token: "entity.other.attribute-name.scss", foreground: "50fa7b" },
      { token: "keyword", foreground: "ff79c6" },
      { token: "storage", foreground: "8be9fd" },
      { token: "storage.type", foreground: "8be9fd", fontStyle: "italic" },
      { token: "string", foreground: "f1fa8c" },
      { token: "string.template", foreground: "f1fa8c" },
      { token: "punctuation.definition.string", foreground: "f1fa8c" },
      { token: "variable.parameter", foreground: "ffb86c", fontStyle: "italic" },
      { token: "variable.name", foreground: "f8f8f2" },
      { token: "support.function", foreground: "8be9fd" },
      { token: "support.variable", foreground: "8be9fd" },
      { token: "support.type", foreground: "8be9fd" },
      { token: "support.class", foreground: "8be9fd" },
      { token: "support.constant", foreground: "bd93f9" },
      { token: "support.other", foreground: "f8f8f2" },
      { token: "meta.function.js", foreground: "8be9fd" },
      { token: "meta.function.method.js", foreground: "8be9fd" },
      { token: "meta.object-literal.key.js", foreground: "f8f8f2" },
      { token: "meta.property.object.js", foreground: "f8f8f2" },
      { token: "variable.other.object.js", foreground: "f8f8f2" },
      { token: "variable.other.property.js", foreground: "f8f8f2" },
      { token: "variable.js", foreground: "f8f8f2" },
      { token: "entity.name.function.js", foreground: "50fa7b" },
      { token: "entity.name.function.method.js", foreground: "50fa7b" },
      { token: "entity.name.type.js", foreground: "8be9fd" },
      { token: "entity.name.class.js", foreground: "8be9fd" },
      { token: "entity.name.tag.js", foreground: "ff79c6" },
      { token: "entity.other.attribute-name.js", foreground: "50fa7b" },
    ],
    colors: {
      "editor.foreground": "#f8f8f2",
      "editor.background": "#282a36",
      "editor.selectionBackground": "#44475a",
      "editor.lineHighlightBackground": "#44475a",
      "editorCursor.foreground": "#f8f8f2",
      "editorWhitespace.foreground": "#6272a4",
      "editorIndentGuide.background": "#6272a4",
      "editorIndentGuide.activeBackground": "#f8f8f2",
    },
  });
}
