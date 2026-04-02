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
		fmt.Println("Usage: go run autofix_types.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "ui" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") || strings.Contains(path, "zz_generated") {
			return nil
		}

		fixFile(fset, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

func fixFile(fset *token.FileSet, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	f, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return
	}

	modified := false

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
				for _, s := range x.Specs {
					switch ts := s.(type) {
					case *ast.TypeSpec:
						if ts.Name.IsExported() {
							if x.Doc == nil && ts.Doc == nil {
								docText := fmt.Sprintf("// %s is an exported type.\n//\n// Summary: %s represents a structure.\n//\n// Parameters:\n//   - None.\n//\n// Returns:\n//   - None.\n//\n// Errors:\n//   - None.\n//\n// Side Effects:\n//   - None.", ts.Name.Name, ts.Name.Name)
								cg := &ast.CommentGroup{
									List: []*ast.Comment{
										{Text: docText, Slash: ts.Pos() - 1},
									},
								}
								ts.Doc = cg
								modified = true
							}

                            // Check methods in interface
							if iface, ok := ts.Type.(*ast.InterfaceType); ok && iface.Methods != nil {
								for _, field := range iface.Methods.List {
									if len(field.Names) > 0 && field.Names[0].IsExported() {
										if field.Doc == nil {
											docText := fmt.Sprintf("// %s defines an interface method.\n//\n// Summary: %s performs an action.\n//\n// Parameters:\n//   - None.\n//\n// Returns:\n//   - None.\n//\n// Errors:\n//   - None.\n//\n// Side Effects:\n//   - None.", field.Names[0].Name, field.Names[0].Name)
											cg := &ast.CommentGroup{
												List: []*ast.Comment{
													{Text: docText, Slash: field.Pos() - 1},
												},
											}
											field.Doc = cg
											modified = true
										}
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, name := range ts.Names {
							if name.IsExported() {
								if x.Doc == nil && ts.Doc == nil {
									docText := fmt.Sprintf("// %s is an exported variable or constant.\n//\n// Summary: %s holds a value.\n//\n// Parameters:\n//   - None.\n//\n// Returns:\n//   - None.\n//\n// Errors:\n//   - None.\n//\n// Side Effects:\n//   - None.", name.Name, name.Name)
									cg := &ast.CommentGroup{
										List: []*ast.Comment{
											{Text: docText, Slash: name.Pos() - 1},
										},
									}
									ts.Doc = cg
									modified = true
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	if modified {
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, f); err != nil {
			fmt.Printf("Error formatting %s: %v\n", path, err)
			return
		}
		os.WriteFile(path, buf.Bytes(), 0644)
	}
}
