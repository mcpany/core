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
		fmt.Println("Usage: go run update_docs_ast.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "test" || info.Name() == "mocks" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		modified := processFile(f, fset, path)
		if modified {
			var buf bytes.Buffer
			if err := format.Node(&buf, fset, f); err != nil {
				fmt.Printf("Error formatting %s: %v\n", path, err)
				return nil
			}

			// Use simple string replacement for existing comments if ast modification fails
			// But AST modification is much safer
			err = os.WriteFile(path, buf.Bytes(), 0644)
			if err != nil {
				fmt.Printf("Error writing %s: %v\n", path, err)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

func processFile(f *ast.File, fset *token.FileSet, path string) bool {
	modified := false

	extractFields := func(fl *ast.FieldList) []string {
		if fl == nil {
			return nil
		}
		var res []string
		for _, field := range fl.List {
			var typeName string
			switch t := field.Type.(type) {
			case *ast.Ident:
				typeName = t.Name
			case *ast.StarExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					typeName = "*" + ident.Name
				} else if sel, ok := t.X.(*ast.SelectorExpr); ok {
					if identX, ok := sel.X.(*ast.Ident); ok {
						typeName = "*" + identX.Name + "." + sel.Sel.Name
					}
				}
			case *ast.SelectorExpr:
				if identX, ok := t.X.(*ast.Ident); ok {
					typeName = identX.Name + "." + t.Sel.Name
				} else {
					typeName = "unknown"
				}
			case *ast.ArrayType:
				typeName = "array/slice"
			case *ast.MapType:
				typeName = "map"
			case *ast.InterfaceType:
				typeName = "interface"
			case *ast.FuncType:
				typeName = "func"
			default:
				typeName = "unknown"
			}

			if len(field.Names) == 0 {
				res = append(res, fmt.Sprintf("unnamed (%s): description", typeName))
			} else {
				for _, name := range field.Names {
					res = append(res, fmt.Sprintf("%s (%s): description", name.Name, typeName))
				}
			}
		}
		return res
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				params := extractFields(x.Type.Params)
				returns := extractFields(x.Type.Results)

				if x.Doc == nil {
					x.Doc = &ast.CommentGroup{
						List: []*ast.Comment{
							{Text: fmt.Sprintf("// Summary: %s ...", x.Name.Name)},
						},
					}
					modified = true
				}
				newDoc, mod := updateDocString(x.Doc.Text(), x.Name.Name, "function", params, returns)
				if mod {
					x.Doc.List = makeComments(newDoc)
					modified = true
				}
			}
		case *ast.GenDecl:
			if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
				for _, s := range x.Specs {
					switch ts := s.(type) {
					case *ast.TypeSpec:
						if ts.Name.IsExported() {
							doc := x.Doc
							if doc == nil {
								doc = ts.Doc
							}
							if doc == nil {
								doc = &ast.CommentGroup{
									List: []*ast.Comment{
										{Text: fmt.Sprintf("// Summary: %s ...", ts.Name.Name)},
									},
								}
								x.Doc = doc
								modified = true
							}
							newDoc, mod := updateDocString(doc.Text(), ts.Name.Name, "type", nil, nil)
							if mod {
								doc.List = makeComments(newDoc)
								modified = true
							}
						}
					case *ast.ValueSpec:
						for _, name := range ts.Names {
							if name.IsExported() {
								doc := x.Doc
								if doc == nil {
									doc = ts.Doc
								}
								if doc == nil {
									doc = &ast.CommentGroup{
										List: []*ast.Comment{
											{Text: fmt.Sprintf("// Summary: %s ...", name.Name)},
										},
									}
									if len(ts.Names) == 1 {
										ts.Doc = doc
									} else {
										x.Doc = doc
									}
									modified = true
								}
								newDoc, mod := updateDocString(doc.Text(), name.Name, "variable/constant", nil, nil)
								if mod {
									doc.List = makeComments(newDoc)
									modified = true
								}
								break
							}
						}
					}
				}
			}
		}
		return true
	})

	return modified
}

func updateDocString(text, name, kind string, params, returns []string) (string, bool) {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	hasSummary := false
	hasParams := false
	hasReturns := false
	hasErrors := false
	hasSideEffects := false

	for _, line := range lines {
		if strings.Contains(line, "Summary:") { hasSummary = true }
		if strings.Contains(line, "Parameters:") { hasParams = true }
		if strings.Contains(line, "Returns:") { hasReturns = true }
		if strings.Contains(line, "Errors:") { hasErrors = true }
		if strings.Contains(line, "Side Effects:") { hasSideEffects = true }
	}

	var newLines []string
	modified := false

	if !hasSummary {
		if len(lines) > 0 && lines[0] != "" {
			if !strings.HasPrefix(lines[0], "Summary:") {
				newLines = append(newLines, "Summary: " + lines[0])
				for i := 1; i < len(lines); i++ {
					newLines = append(newLines, lines[i])
				}
				modified = true
			} else {
				newLines = append(newLines, lines...)
			}
		} else {
			newLines = append(newLines, fmt.Sprintf("Summary: %s is a %s.", name, kind))
			modified = true
		}
	} else {
		newLines = append(newLines, lines...)
	}

	if kind == "function" {
		if !hasParams {
			newLines = append(newLines, "")
			newLines = append(newLines, "Parameters:")
			if len(params) == 0 {
				newLines = append(newLines, "  - None.")
			} else {
				for _, p := range params {
					newLines = append(newLines, fmt.Sprintf("  - %s", p))
				}
			}
			modified = true
		}

		if !hasReturns {
			newLines = append(newLines, "")
			newLines = append(newLines, "Returns:")
			if len(returns) == 0 {
				newLines = append(newLines, "  - None.")
			} else {
				for _, r := range returns {
					newLines = append(newLines, fmt.Sprintf("  - %s", r))
				}
			}
			modified = true
		}

		if !hasErrors {
			newLines = append(newLines, "")
			newLines = append(newLines, "Errors:")
			newLines = append(newLines, "  - None.")
			modified = true
		}

		if !hasSideEffects {
			newLines = append(newLines, "")
			newLines = append(newLines, "Side Effects:")
			newLines = append(newLines, "  - None.")
			modified = true
		}
	} else {
		if !hasSideEffects {
			newLines = append(newLines, "")
			newLines = append(newLines, "Side Effects:")
			newLines = append(newLines, "  - None.")
			modified = true
		}
	}

	return strings.Join(newLines, "\n"), modified
}

func makeComments(text string) []*ast.Comment {
	lines := strings.Split(text, "\n")
	var comments []*ast.Comment
	for _, l := range lines {
		comments = append(comments, &ast.Comment{Text: "// " + l})
	}
	return comments
}
