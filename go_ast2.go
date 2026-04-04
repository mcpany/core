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
	fset := token.NewFileSet()
	var missing []string

	err := filepath.Walk("server/pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "proto") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					doc := ""
                    if d.Doc != nil {
                        doc = d.Doc.Text()
                    }
					if !strings.Contains(doc, "Summary:") {
						missing = append(missing, fmt.Sprintf("%s: func %s", path, d.Name.Name))
					}
				}
			case *ast.GenDecl:
				if d.Tok == token.TYPE || d.Tok == token.CONST || d.Tok == token.VAR {
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() {
                                doc := ""
                                if d.Doc != nil {
                                    doc = d.Doc.Text()
                                }
                                if s.Doc != nil {
                                    doc += s.Doc.Text()
                                }
								if !strings.Contains(doc, "Summary:") {
									missing = append(missing, fmt.Sprintf("%s: type %s", path, s.Name.Name))
								}
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
                                    doc := ""
                                    if d.Doc != nil {
                                        doc = d.Doc.Text()
                                    }
                                    if s.Doc != nil {
                                        doc += s.Doc.Text()
                                    }
									if !strings.Contains(doc, "Summary:") {
										missing = append(missing, fmt.Sprintf("%s: var/const %s", path, name.Name))
									}
								}
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, m := range missing {
		fmt.Println(m)
	}
	fmt.Printf("Total missing in server/pkg: %d\n", len(missing))
}
