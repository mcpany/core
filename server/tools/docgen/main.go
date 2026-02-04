// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run server/tools/docgen/main.go <directory>...")
		os.Exit(1)
	}

	dirs := os.Args[1:]
	for _, dir := range dirs {
		if err := filepath.Walk(dir, walkFunc); err != nil {
			fmt.Printf("Error walking %s: %v\n", dir, err)
		}
	}
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "testdata" {
			return filepath.SkipDir
		}
		return nil
	}
	if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
		return nil
	}

	return processFile(path)
}

func processFile(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	modified := false
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if updateFuncDoc(x) {
					modified = true
				}
			}
		case *ast.GenDecl:
			if updateGenDeclDoc(x) {
				modified = true
			}
		}
		return true
	})

	if modified {
		fmt.Printf("Updating %s\n", path)
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, f); err != nil {
			return fmt.Errorf("formatting %s: %w", path, err)
		}
		return os.WriteFile(path, buf.Bytes(), 0644)
	}
	return nil
}

func updateFuncDoc(fn *ast.FuncDecl) bool {
	if fn.Doc != nil {
		text := fn.Doc.Text()
		if strings.Contains(text, "Summary:") && strings.Contains(text, "Parameters:") {
			return false
		}
	}

	// Extract existing summary
	summary := extractSummary(fn.Doc, fn.Name.Name)

	// Build new doc comment
	var newDoc strings.Builder
	newDoc.WriteString(fmt.Sprintf("// %s %s\n", fn.Name.Name, summary))
	newDoc.WriteString("//\n")
	newDoc.WriteString(fmt.Sprintf("// Summary: %s\n", summary))

	// Parameters
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		newDoc.WriteString("//\n")
		newDoc.WriteString("// Parameters:\n")
		for _, field := range fn.Type.Params.List {
			typeStr := exprToString(field.Type)
			for _, name := range field.Names {
				newDoc.WriteString(fmt.Sprintf("//   - %s (%s): %s\n", name.Name, typeStr, inferParamDesc(name.Name, typeStr)))
			}
			if len(field.Names) == 0 {
				newDoc.WriteString(fmt.Sprintf("//   - _ (%s): Parameter\n", typeStr))
			}
		}
	}

	// Returns
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		newDoc.WriteString("//\n")
		newDoc.WriteString("// Returns:\n")
		for _, field := range fn.Type.Results.List {
			typeStr := exprToString(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					newDoc.WriteString(fmt.Sprintf("//   - %s (%s): %s\n", name.Name, typeStr, inferReturnDesc(name.Name, typeStr)))
				}
			} else {
				newDoc.WriteString(fmt.Sprintf("//   - %s: %s\n", typeStr, inferReturnDesc("", typeStr)))
			}
		}
	}

	details := extractDetails(fn.Doc)
	if details != "" {
		newDoc.WriteString("//\n")
		if !strings.Contains(details, "Summary:") {
			newDoc.WriteString("// Description:\n")
			for _, line := range strings.Split(details, "\n") {
				newDoc.WriteString(fmt.Sprintf("// %s\n", line))
			}
		}
	}

	comments := []*ast.Comment{}
	for _, line := range strings.Split(strings.TrimSuffix(newDoc.String(), "\n"), "\n") {
		comments = append(comments, &ast.Comment{Text: line})
	}

	fn.Doc = &ast.CommentGroup{List: comments}
	return true
}

func updateGenDeclDoc(gen *ast.GenDecl) bool {
	if gen.Tok == token.TYPE {
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if !ts.Name.IsExported() {
				continue
			}

			targetDoc := gen.Doc
			if gen.Lparen.IsValid() {
				targetDoc = ts.Doc
			}

			if targetDoc != nil {
				text := targetDoc.Text()
				if strings.Contains(text, "Summary:") {
					continue
				}
			}

			summary := extractSummary(targetDoc, ts.Name.Name)

			var newDoc strings.Builder
			newDoc.WriteString(fmt.Sprintf("// %s %s\n", ts.Name.Name, summary))
			newDoc.WriteString("//\n")
			newDoc.WriteString(fmt.Sprintf("// Summary: %s\n", summary))

			comments := []*ast.Comment{}
			for _, line := range strings.Split(strings.TrimSuffix(newDoc.String(), "\n"), "\n") {
				comments = append(comments, &ast.Comment{Text: line})
			}

			if gen.Lparen.IsValid() {
				ts.Doc = &ast.CommentGroup{List: comments}
			} else {
				gen.Doc = &ast.CommentGroup{List: comments}
			}
			return true
		}
	}
	return false
}

func extractSummary(group *ast.CommentGroup, name string) string {
	if group == nil {
		return generateSummaryFromName(name)
	}
	text := group.Text()
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[0] != "" {
		clean := strings.TrimSpace(lines[0])
		if strings.HasPrefix(clean, name) {
			clean = strings.TrimSpace(strings.TrimPrefix(clean, name))
		}
		if clean != "" {
			return clean
		}
	}
	return generateSummaryFromName(name)
}

func generateSummaryFromName(name string) string {
	// Simple camelCase splitter
	var words []string
	start := 0
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			words = append(words, name[start:i])
			start = i
		}
	}
	words = append(words, name[start:])

	// Convert to sentence
	var sentence strings.Builder
	if len(words) > 0 {
		// Heuristics for verbs
		verb := strings.ToLower(words[0])
		switch verb {
		case "get":
			sentence.WriteString("Retrieves the ")
		case "set":
			sentence.WriteString("Sets the ")
		case "new":
			sentence.WriteString("Creates a new ")
		case "validate":
			sentence.WriteString("Validates the ")
		case "update":
			sentence.WriteString("Updates the ")
		case "create":
			sentence.WriteString("Creates the ")
		case "delete":
			sentence.WriteString("Deletes the ")
		case "is", "has", "can":
			sentence.WriteString("Checks if ")
		default:
			// Just use the word
			sentence.WriteString(words[0] + " ")
		}

		for i := 1; i < len(words); i++ {
			sentence.WriteString(strings.ToLower(words[i]) + " ")
		}
	}

	res := strings.TrimSpace(sentence.String())
	if !strings.HasSuffix(res, ".") {
		res += "."
	}
	return res
}

func extractDetails(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	text := group.Text()
	parts := strings.SplitN(text, "\n", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", exprToString(t.Key), exprToString(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	default:
		return "any"
	}
}

func inferParamDesc(name, typeStr string) string {
	name = strings.ToLower(name)
	if name == "ctx" {
		return "Controls the lifecycle of the operation (cancellation, timeouts)."
	}
	if name == "w" && typeStr == "http.ResponseWriter" {
		return "Writes the HTTP response."
	}
	if name == "r" && typeStr == "*http.Request" {
		return "Represents the incoming HTTP request."
	}
	if strings.Contains(name, "id") {
		return "Unique identifier for the entity."
	}
	if strings.Contains(name, "name") {
		return "Name of the resource."
	}
	if strings.Contains(name, "config") || strings.Contains(name, "cfg") {
		return "Configuration parameters."
	}
	if strings.Contains(name, "client") {
		return "Client instance for external communication."
	}
	return fmt.Sprintf("The %s parameter.", name)
}

func inferReturnDesc(name, typeStr string) string {
	if typeStr == "error" {
		return "An error if the operation failed, or nil on success."
	}
	if typeStr == "bool" {
		return "True if the condition is met, false otherwise."
	}
	if strings.HasPrefix(typeStr, "*") {
		return fmt.Sprintf("A pointer to %s, or nil if not found/failed.", typeStr[1:])
	}
	return fmt.Sprintf("The resulting %s.", typeStr)
}
