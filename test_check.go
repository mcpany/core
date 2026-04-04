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
	root := "src"
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		checkFile(f, fset, path)
		return nil
	})

	if err != nil {
		os.Exit(1)
	}
}

func hasRequiredDocs(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	text := doc.Text()
	return strings.Contains(text, "Summary:") &&
		strings.Contains(text, "Parameters:") &&
		strings.Contains(text, "Returns:") &&
		(strings.Contains(text, "Errors:") || strings.Contains(text, "Throws/Errors:")) &&
		strings.Contains(text, "Side Effects:")
}

func checkFile(f *ast.File, fset *token.FileSet, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if !hasRequiredDocs(x.Doc) {
					fmt.Printf("%s:%d: missing or incomplete doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
				}
			}
		case *ast.GenDecl:
			if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
				for _, s := range x.Specs {
					switch ts := s.(type) {
					case *ast.TypeSpec:
						if ts.Name.IsExported() {
							if !hasRequiredDocs(x.Doc) && !hasRequiredDocs(ts.Doc) {
								fmt.Printf("%s:%d: missing or incomplete doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
							}
							if iface, ok := ts.Type.(*ast.InterfaceType); ok {
								for _, field := range iface.Methods.List {
									if len(field.Names) > 0 && field.Names[0].IsExported() {
										if !hasRequiredDocs(field.Doc) {
											fmt.Printf("%s:%d: missing or incomplete doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
										}
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, name := range ts.Names {
							if name.IsExported() {
								if !hasRequiredDocs(x.Doc) && !hasRequiredDocs(ts.Doc) {
									fmt.Printf("%s:%d: missing or incomplete doc for var/const %s\n", path, fset.Position(name.Pos()).Line, name.Name)
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}
