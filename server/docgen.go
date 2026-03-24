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
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run docgen.go <directory>")
		os.Exit(1)
	}
	dir := os.Args[1]

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") && !strings.HasPrefix(filepath.Base(path), "mock_") {
			processFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func hasStructuredDoc(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	text := doc.Text()
	return strings.Contains(text, "Summary:") && strings.Contains(text, "Parameters:") && strings.Contains(text, "Returns:") && strings.Contains(text, "Errors:") && strings.Contains(text, "Side Effects:")
}

func cleanDocText(text string) string {
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "Parameters:") && !strings.HasPrefix(l, "Returns:") && !strings.HasPrefix(l, "Errors:") && !strings.HasPrefix(l, "Side Effects:") && !strings.HasPrefix(l, "- None") && !strings.HasPrefix(l, "- TODO:") && !strings.HasPrefix(l, "- error:") {
			cleaned = append(cleaned, l)
		}
	}
	return strings.Join(cleaned, " ")
}

func formatParam(p string) string {
	parts := strings.SplitN(p, " ", 2)
	if len(parts) == 2 {
		return fmt.Sprintf("//   - %s (%s): The %s parameter.", parts[0], parts[1], parts[0])
	}
	return fmt.Sprintf("//   - %s: The %s parameter.", parts[0], parts[0])
}

func formatReturn(r string) string {
	parts := strings.SplitN(r, " ", 2)

	if len(parts) == 2 {
		return fmt.Sprintf("//   - %s: The resulting %s.", parts[1], parts[1])
	}
	return fmt.Sprintf("//   - %s: The resulting %s.", parts[0], parts[0])
}

func generateSummary(name string, kind string) string {
    if kind == "func" {
        if strings.HasPrefix(name, "New") {
            return fmt.Sprintf("Initializes a new %s.", strings.TrimPrefix(name, "New"))
        }
        if strings.HasPrefix(name, "Get") {
            return fmt.Sprintf("Retrieves the %s.", strings.TrimPrefix(name, "Get"))
        }
        if strings.HasPrefix(name, "Set") {
            return fmt.Sprintf("Sets the %s.", strings.TrimPrefix(name, "Set"))
        }
        if strings.HasPrefix(name, "Is") || strings.HasPrefix(name, "Has") {
            return fmt.Sprintf("Checks if %s.", name)
        }
        return fmt.Sprintf("Executes the %s operation.", name)
    } else if kind == "type" {
        return fmt.Sprintf("Represents the %s type.", name)
    } else {
        return fmt.Sprintf("Defines the %s constant/variable.", name)
    }
}

func createStructuredDoc(name string, params, returns []string, existingText string, kind string) []string {
	var lines []string

	summary := ""
	if existingText != "" {
		linesText := strings.Split(existingText, "\n")
		for _, l := range linesText {
			if strings.HasPrefix(l, "Summary:") {
				summary = strings.TrimSpace(strings.TrimPrefix(l, "Summary:"))
				break
			}
		}
		if summary == "" {
			summary = cleanDocText(existingText)
		}
	}

	if summary == "" || summary == name {
        summary = generateSummary(name, kind)
	}

	if !strings.HasPrefix(summary, name) {
		if len(summary) > 0 && unicode.IsUpper(rune(summary[0])) {
            // keep it if upper
		} else {
            summary = name + " " + summary
        }
	}

	lines = append(lines, fmt.Sprintf("// Summary: %s", summary))
	lines = append(lines, "//")

	lines = append(lines, "// Parameters:")
	if len(params) > 0 {
		for _, p := range params {
			lines = append(lines, formatParam(p))
		}
	} else {
		lines = append(lines, "//   - None.")
	}
	lines = append(lines, "//")

	lines = append(lines, "// Returns:")
	hasError := false
	if len(returns) > 0 {
		for _, r := range returns {
			if r == "error" || strings.HasSuffix(r, " error") {
				hasError = true
				lines = append(lines, "//   - error: An error if the operation fails.")
			} else {
				lines = append(lines, formatReturn(r))
			}
		}
	} else {
		lines = append(lines, "//   - None.")
	}
	lines = append(lines, "//")

	lines = append(lines, "// Errors:")
	if hasError {
		lines = append(lines, "//   - Returns an error if the operation fails or is invalid.")
	} else {
		lines = append(lines, "//   - None.")
	}
	lines = append(lines, "//")

	lines = append(lines, "// Side Effects:")
	lines = append(lines, "//   - None.")

	return lines
}

func getFieldListStrings(fl *ast.FieldList, fset *token.FileSet) []string {
	if fl == nil {
		return nil
	}
	var res []string
	for _, field := range fl.List {
		var typeBuf bytes.Buffer
		format.Node(&typeBuf, fset, field.Type)
		typeStr := typeBuf.String()

		if len(field.Names) == 0 {
			res = append(res, typeStr)
		} else {
			for _, name := range field.Names {
				res = append(res, name.Name+" "+typeStr)
			}
		}
	}
	return res
}

type insertOp struct {
	startLine int
	endLine   int
	docLines  []string
}

func insertDocBefore(fileContent []byte, ops []insertOp) []byte {
	lines := strings.Split(string(fileContent), "\n")

	// Apply operations backwards
	for _, op := range ops {
		var newLines []string
		newLines = append(newLines, lines[:op.startLine]...)
		newLines = append(newLines, op.docLines...)
		newLines = append(newLines, lines[op.endLine:]...)
		lines = newLines
	}

	return []byte(strings.Join(lines, "\n"))
}

func extractDocText(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	text := doc.Text()

	lines := strings.Split(text, "\n")
	var summaryLines []string

	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "Parameters:" || l == "Returns:" || l == "Errors:" || l == "Side Effects:" {
			break
		}
		if l != "" {
			if strings.HasPrefix(l, "Summary:") {
				summaryLines = append(summaryLines, strings.TrimSpace(strings.TrimPrefix(l, "Summary:")))
			} else {
				summaryLines = append(summaryLines, l)
			}
		}
	}

	return strings.Join(summaryLines, " ")
}

func processFile(filepath string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	var ops []insertOp

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				if !hasStructuredDoc(d.Doc) || strings.Contains(d.Doc.Text(), "TODO:") {
					params := getFieldListStrings(d.Type.Params, fset)
					returns := getFieldListStrings(d.Type.Results, fset)

					existingText := extractDocText(d.Doc)
					startLine := fset.Position(d.Pos()).Line - 1
					endLine := startLine
					if d.Doc != nil {
						startLine = fset.Position(d.Doc.Pos()).Line - 1
					}

					docLines := createStructuredDoc(d.Name.Name, params, returns, existingText, "func")
					ops = append(ops, insertOp{startLine: startLine, endLine: endLine, docLines: docLines})
				}
			}
		case *ast.GenDecl:
			if d.Tok == token.TYPE || d.Tok == token.CONST || d.Tok == token.VAR {
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							doc := s.Doc
							if doc == nil && len(d.Specs) == 1 {
								doc = d.Doc
							}

							if !hasStructuredDoc(doc) || strings.Contains(doc.Text(), "TODO:") {
								existingText := extractDocText(doc)
								startLine := fset.Position(s.Pos()).Line - 1
								endLine := startLine
								if len(d.Specs) == 1 {
									startLine = fset.Position(d.Pos()).Line - 1
									endLine = startLine
								}
								if doc != nil {
									startLine = fset.Position(doc.Pos()).Line - 1
								}
								docLines := createStructuredDoc(s.Name.Name, nil, nil, existingText, "type")
								ops = append(ops, insertOp{startLine: startLine, endLine: endLine, docLines: docLines})
							}
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.IsExported() {
								doc := s.Doc
								if doc == nil && len(d.Specs) == 1 {
									doc = d.Doc
								}

								if !hasStructuredDoc(doc) || strings.Contains(doc.Text(), "TODO:") {
									existingText := extractDocText(doc)
									startLine := fset.Position(s.Pos()).Line - 1
									endLine := startLine
									if len(d.Specs) == 1 {
										startLine = fset.Position(d.Pos()).Line - 1
										endLine = startLine
									}
									if doc != nil {
										startLine = fset.Position(doc.Pos()).Line - 1
									}

									kind := "const"
									if d.Tok == token.VAR {
										kind = "var"
									}
									docLines := createStructuredDoc(name.Name, nil, nil, existingText, kind)
									ops = append(ops, insertOp{startLine: startLine, endLine: endLine, docLines: docLines})
								}
								break // only do first exported name
							}
						}
					}
				}
			}
		}
	}

	if len(ops) > 0 {
		for i := 0; i < len(ops); i++ {
			for j := i + 1; j < len(ops); j++ {
				if ops[i].startLine < ops[j].startLine {
					ops[i], ops[j] = ops[j], ops[i]
				}
			}
		}

		fileContent = insertDocBefore(fileContent, ops)
		os.WriteFile(filepath, fileContent, 0644)
		fmt.Println("Updated", filepath)
	}
}
