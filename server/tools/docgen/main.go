// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/fix_docs.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "test" || info.Name() == "tools" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		// Read file content first
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		f, err := parser.ParseFile(fset, path, content, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		newContent := processFile(f, fset, content)
		if newContent != nil {
			if err := os.WriteFile(path, newContent, info.Mode()); err != nil {
				fmt.Printf("Error writing %s: %v\n", path, err)
			} else {
				fmt.Printf("Updated %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

type Replacement struct {
	Start int
	End   int
	Text  string
}

func processFile(f *ast.File, fset *token.FileSet, content []byte) []byte {
	var replacements []Replacement

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if needsDocs(x.Doc) {
					start, end, newDoc := generateDoc(x, content, fset)
					replacements = append(replacements, Replacement{Start: start, End: end, Text: newDoc})
				}
			}
		case *ast.GenDecl:
			// Handle types (structs, interfaces)
			if x.Tok == token.TYPE {
				for _, s := range x.Specs {
					if ts, ok := s.(*ast.TypeSpec); ok {
						if ts.Name.IsExported() {
							// Check if it's the only spec and the doc is on the GenDecl
							var doc *ast.CommentGroup
							if len(x.Specs) == 1 {
								doc = x.Doc
							} else {
								doc = ts.Doc
							}

							if needsDocs(doc) {
								if len(x.Specs) == 1 {
									start, end, newDoc := generateTypeDoc(ts, x.Doc, x, content, fset)
									replacements = append(replacements, Replacement{Start: start, End: end, Text: newDoc})
								}
							}

							// Check methods in interface
							if iface, ok := ts.Type.(*ast.InterfaceType); ok {
								for _, field := range iface.Methods.List {
									if len(field.Names) > 0 && field.Names[0].IsExported() {
										if needsDocs(field.Doc) {
											start, end, newDoc := generateInterfaceMethodDoc(field, content, fset)
											replacements = append(replacements, Replacement{Start: start, End: end, Text: newDoc})
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	if len(replacements) == 0 {
		return nil
	}

	// Sort replacements by Start descending
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].Start > replacements[j].Start
	})

	newContent := make([]byte, len(content))
	copy(newContent, content)

	for _, r := range replacements {
		if r.Start > len(newContent) || r.End > len(newContent) {
			continue
		}

		prefix := newContent[:r.Start]
		suffix := newContent[r.End:]

		var buf bytes.Buffer
		buf.Write(prefix)
		buf.WriteString(r.Text)
		buf.Write(suffix)
		newContent = buf.Bytes()
	}

	return newContent
}

func needsDocs(doc *ast.CommentGroup) bool {
	if doc == nil {
		return true
	}
	text := doc.Text()

	return !strings.Contains(text, "Summary:")
}

func generateDoc(fn *ast.FuncDecl, content []byte, fset *token.FileSet) (int, int, string) {
	name := fn.Name.Name

	startPos := fset.Position(fn.Pos()).Offset
	lineStart := startPos
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	indent := string(content[lineStart:startPos])

	existingSummary := ""
	if fn.Doc != nil {
		existingSummary = fn.Doc.Text()
		existingSummary = strings.TrimSpace(strings.Split(existingSummary, "\n")[0])
		if strings.HasPrefix(existingSummary, name) {
			existingSummary = strings.TrimSpace(strings.TrimPrefix(existingSummary, name))
		}
	}

	if existingSummary == "" {
		splitName := camelCaseSplit(name)
		summary := strings.Join(splitName, " ")
		if strings.HasPrefix(name, "New") {
			summary = "Creates a new " + strings.TrimPrefix(name, "New")
		} else if strings.HasPrefix(name, "Get") {
			summary = "Retrieves the " + strings.TrimPrefix(name, "Get")
		} else if strings.HasPrefix(name, "Set") {
			summary = "Sets the " + strings.TrimPrefix(name, "Set")
		} else {
			summary = summary + " operation"
		}
		existingSummary = summary
	}

	if !strings.HasSuffix(existingSummary, ".") {
		existingSummary += "."
	}

	var sb strings.Builder
	first := true
	writeLine := func(s string) {
		if first {
			sb.WriteString(s + "\n")
			first = false
		} else {
			sb.WriteString(indent + s + "\n")
		}
	}

	writeLine(fmt.Sprintf("// %s %s", name, lowerFirst(existingSummary)))
	writeLine("//")
	writeLine(fmt.Sprintf("// Summary: %s", existingSummary))
	writeLine("//")

	// Parameters
	params := []string{}
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			typeStr := getTypeString(p.Type, content, fset)
			for _, n := range p.Names {
				desc := getDescriptionForParam(n.Name, typeStr)
				params = append(params, fmt.Sprintf("//   - %s: %s. %s.", n.Name, typeStr, desc))
			}
			if len(p.Names) == 0 {
				desc := getDescriptionForParam("", typeStr)
				params = append(params, fmt.Sprintf("//   - <unnamed>: %s. %s.", typeStr, desc))
			}
		}
	}

	if len(params) > 0 {
		writeLine("// Parameters:")
		for _, p := range params {
			writeLine(p)
		}
	} else {
		writeLine("// Parameters:")
		writeLine("//   None.")
	}

	// Returns
	returns := []string{}
	if fn.Type.Results != nil {
		for _, r := range fn.Type.Results.List {
			typeStr := getTypeString(r.Type, content, fset)
			desc := fmt.Sprintf("The %s", typeStr)
			if typeStr == "error" {
				desc = "An error if the operation fails"
			}
			if len(r.Names) > 0 {
				for _, n := range r.Names {
					returns = append(returns, fmt.Sprintf("//   - %s: %s. %s.", n.Name, typeStr, desc))
				}
			} else {
				returns = append(returns, fmt.Sprintf("//   - %s: %s.", typeStr, desc))
			}
		}
	}

	writeLine("//")
	writeLine("// Returns:")
	if len(returns) > 0 {
		for _, r := range returns {
			writeLine(r)
		}
	} else {
		writeLine("//   None.")
	}

	hasError := false
	if fn.Type.Results != nil {
		for _, r := range fn.Type.Results.List {
			if getTypeString(r.Type, content, fset) == "error" {
				hasError = true
				break
			}
		}
	}

	if hasError {
		writeLine("//")
		writeLine("// Throws/Errors:")
		writeLine("//   Returns an error if the operation fails.")
	}

	start := startPos
	end := start // insert mode if no doc

	if fn.Doc != nil {
		start = fset.Position(fn.Doc.Pos()).Offset
		end = fset.Position(fn.Doc.End()).Offset
		if end < len(content) && content[end] == '\n' {
			end++
		}
	} else {
		// If inserting, we need to append indent at the end so func starts correctly
		sb.WriteString(indent)
	}

	return start, end, sb.String()
}

func generateTypeDoc(ts *ast.TypeSpec, declDoc *ast.CommentGroup, gd *ast.GenDecl, content []byte, fset *token.FileSet) (int, int, string) {
	name := ts.Name.Name

	// Use GenDecl Pos for indentation
	startPos := fset.Position(gd.Pos()).Offset
	lineStart := startPos
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	indent := string(content[lineStart:startPos])

	existingSummary := ""
	if declDoc != nil {
		existingSummary = declDoc.Text()
		existingSummary = strings.TrimSpace(strings.Split(existingSummary, "\n")[0])
		if strings.HasPrefix(existingSummary, name) {
			existingSummary = strings.TrimSpace(strings.TrimPrefix(existingSummary, name))
		}
	}

	if existingSummary == "" {
		existingSummary = fmt.Sprintf("Represents the %s structure", name)
	}

	if !strings.HasSuffix(existingSummary, ".") {
		existingSummary += "."
	}

	var sb strings.Builder
	first := true
	writeLine := func(s string) {
		if first {
			sb.WriteString(s + "\n")
			first = false
		} else {
			sb.WriteString(indent + s + "\n")
		}
	}

	writeLine(fmt.Sprintf("// %s %s", name, lowerFirst(existingSummary)))
	writeLine("//")
	writeLine(fmt.Sprintf("// Summary: %s", existingSummary))

	start := startPos
	end := start

	if declDoc != nil {
		start = fset.Position(declDoc.Pos()).Offset
		end = fset.Position(declDoc.End()).Offset
		if end < len(content) && content[end] == '\n' {
			end++
		}
	} else {
		sb.WriteString(indent)
	}

	return start, end, sb.String()
}

func generateInterfaceMethodDoc(field *ast.Field, content []byte, fset *token.FileSet) (int, int, string) {
	name := field.Names[0].Name

	startPos := fset.Position(field.Pos()).Offset
	lineStart := startPos
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	indent := string(content[lineStart:startPos])

	existingSummary := ""
	if field.Doc != nil {
		existingSummary = field.Doc.Text()
		existingSummary = strings.TrimSpace(strings.Split(existingSummary, "\n")[0])
		if strings.HasPrefix(existingSummary, name) {
			existingSummary = strings.TrimSpace(strings.TrimPrefix(existingSummary, name))
		}
	}

	if existingSummary == "" {
		splitName := camelCaseSplit(name)
		summary := strings.Join(splitName, " ")
		summary = summary + " operation"
		existingSummary = summary
	}

	if !strings.HasSuffix(existingSummary, ".") {
		existingSummary += "."
	}

	var sb strings.Builder
	first := true
	writeLine := func(s string) {
		if first {
			sb.WriteString(s + "\n")
			first = false
		} else {
			sb.WriteString(indent + s + "\n")
		}
	}

	writeLine(fmt.Sprintf("// %s %s", name, lowerFirst(existingSummary)))
	writeLine("//")
	writeLine(fmt.Sprintf("// Summary: %s", existingSummary))
	writeLine("//")

	// FuncType
	if ft, ok := field.Type.(*ast.FuncType); ok {
		// Parameters
		params := []string{}
		if ft.Params != nil {
			for _, p := range ft.Params.List {
				typeStr := getTypeString(p.Type, content, fset)
				for _, n := range p.Names {
					desc := getDescriptionForParam(n.Name, typeStr)
					params = append(params, fmt.Sprintf("//   - %s: %s. %s.", n.Name, typeStr, desc))
				}
				if len(p.Names) == 0 {
					desc := getDescriptionForParam("", typeStr)
					params = append(params, fmt.Sprintf("//   - <unnamed>: %s. %s.", typeStr, desc))
				}
			}
		}

		if len(params) > 0 {
			writeLine("// Parameters:")
			for _, p := range params {
				writeLine(p)
			}
		} else {
			writeLine("// Parameters:")
			writeLine("//   None.")
		}

		// Returns
		returns := []string{}
		if ft.Results != nil {
			for _, r := range ft.Results.List {
				typeStr := getTypeString(r.Type, content, fset)
				desc := fmt.Sprintf("The %s", typeStr)
				if typeStr == "error" {
					desc = "An error if the operation fails"
				}
				if len(r.Names) > 0 {
					for _, n := range r.Names {
						returns = append(returns, fmt.Sprintf("//   - %s: %s. %s.", n.Name, typeStr, desc))
					}
				} else {
					returns = append(returns, fmt.Sprintf("//   - %s: %s.", typeStr, desc))
				}
			}
		}

		writeLine("//")
		writeLine("// Returns:")
		if len(returns) > 0 {
			for _, r := range returns {
				writeLine(r)
			}
		} else {
			writeLine("//   None.")
		}

		hasError := false
		if ft.Results != nil {
			for _, r := range ft.Results.List {
				if getTypeString(r.Type, content, fset) == "error" {
					hasError = true
					break
				}
			}
		}

		if hasError {
			writeLine("//")
			writeLine("// Throws/Errors:")
			writeLine("//   Returns an error if the operation fails.")
		}
	}

	start := startPos
	end := start

	if field.Doc != nil {
		start = fset.Position(field.Doc.Pos()).Offset
		end = fset.Position(field.Doc.End()).Offset
		if end < len(content) && content[end] == '\n' {
			end++
		}
	} else {
		sb.WriteString(indent)
	}

	return start, end, sb.String()
}

func getTypeString(expr ast.Expr, content []byte, fset *token.FileSet) string {
	start := fset.Position(expr.Pos()).Offset
	end := fset.Position(expr.End()).Offset
	return string(content[start:end])
}

func camelCaseSplit(s string) []string {
	var parts []string
	var current strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(s) && !unicode.IsUpper(rune(s[i+1]))) {
			parts = append(parts, current.String())
			current.Reset()
		}
		current.WriteRune(r)
	}
	parts = append(parts, current.String())
	return parts
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func getDescriptionForParam(name, typeStr string) string {
	// Check known names
	switch name {
	case "ctx":
		return "The context for the operation"
	case "err":
		return "An error if the operation fails"
	case "req":
		return "The request object"
	case "resp":
		return "The response object"
	case "w":
		return "The HTTP response writer"
	case "r":
		return "The HTTP request"
	case "path":
		return "The file path"
	case "url":
		return "The URL"
	case "id":
		return "The unique identifier"
	case "name":
		return "The name"
	case "config":
		return "The configuration"
	case "options":
		return "The options"
	case "logger":
		return "The logger instance"
	case "command":
		return "The command to execute"
	case "args":
		return "The command arguments"
	case "workingDir":
		return "The working directory"
	case "env":
		return "The environment variables"
	}

	// Check known types
	if strings.HasSuffix(typeStr, "Context") {
		return "The context for the operation"
	}
	if strings.HasSuffix(typeStr, "Request") {
		return "The request object"
	}
	if strings.HasSuffix(typeStr, "Response") {
		return "The response object"
	}
	if typeStr == "error" {
		return "An error if the operation fails"
	}

	// Heuristic: split type name
	// *configv1.ContainerEnvironment -> "The container environment"
	// Clean typeStr: remove *, package name
	cleanType := strings.TrimPrefix(typeStr, "*")
	if idx := strings.LastIndex(cleanType, "."); idx != -1 {
		cleanType = cleanType[idx+1:]
	}
	// Split CamelCase
	parts := camelCaseSplit(cleanType)
	return "The " + strings.ToLower(strings.Join(parts, " "))
}
