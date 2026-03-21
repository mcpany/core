package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_check.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "tests" || info.Name() == "test" || info.Name() == "examples" || info.Name() == "docs" || info.Name() == "cmd" || info.Name() == "api" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		if strings.Contains(path, "mock_") || strings.Contains(path, "mocks") {
		    return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		checkFile(f, fset, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

func checkFile(f *ast.File, fset *token.FileSet, path string) {
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				checkDocstring(x.Doc, fset, path, "function "+x.Name.Name)
			}
		case *ast.GenDecl:
			if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
				for _, s := range x.Specs {
					switch ts := s.(type) {
					case *ast.TypeSpec:
						if ts.Name.IsExported() {
							doc := x.Doc
							if ts.Doc != nil {
								doc = ts.Doc
							}
							checkDocstring(doc, fset, path, "type "+ts.Name.Name)

							// Check methods in interface
							if iface, ok := ts.Type.(*ast.InterfaceType); ok {
								for _, field := range iface.Methods.List {
									if len(field.Names) > 0 && field.Names[0].IsExported() {
										checkDocstring(field.Doc, fset, path, "interface method "+ts.Name.Name+"."+field.Names[0].Name)
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, name := range ts.Names {
							if name.IsExported() {
								doc := x.Doc
								if ts.Doc != nil {
									doc = ts.Doc
								}
								checkDocstring(doc, fset, path, "var/const "+name.Name)
							}
						}
					}
				}
			}
		}
	}
}

func checkDocstring(doc *ast.CommentGroup, fset *token.FileSet, path, name string) {
	if doc == nil {
		fmt.Printf("%s: missing doc for %s\n", path, name)
		return
	}

	text := doc.Text()

	hasSummary := strings.Contains(text, "Summary:")
	hasParams := strings.Contains(text, "Parameters:")
	hasReturns := strings.Contains(text, "Returns:")
	hasErrors := strings.Contains(text, "Errors:")
	hasSideEffects := strings.Contains(text, "Side Effects:")

	if !hasSummary || !hasParams || !hasReturns || !hasErrors || !hasSideEffects {
		fmt.Printf("%s: malformed doc for %s\n", path, name)
	}
}
