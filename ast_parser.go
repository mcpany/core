package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Node struct {
	File     string  `json:"file"`
	Kind     string  `json:"kind"` // func, method, type, const, var
	Name     string  `json:"name"`
	Recv     string  `json:"recv,omitempty"`
	Params   []Field `json:"params,omitempty"`
	Results  []Field `json:"results,omitempty"`
	Doc      string  `json:"doc"`
	Pos      int     `json:"pos"`       // position to insert doc if none exists (1-based line number)
	DocStart int     `json:"doc_start"` // start of existing doc (1-based line number)
	DocEnd   int     `json:"doc_end"`   // end of existing doc (1-based line number)
}

func exprToString(expr ast.Expr, src []byte, fset *token.FileSet) string {
	if expr == nil {
		return ""
	}
	start := fset.Position(expr.Pos()).Offset
	end := fset.Position(expr.End()).Offset
	if start >= 0 && end > start && end <= len(src) {
	    return string(src[start:end])
	}
	return ""
}

func getFields(fl *ast.FieldList, src []byte, fset *token.FileSet) []Field {
	if fl == nil {
		return nil
	}
	var res []Field
	for _, f := range fl.List {
		typ := exprToString(f.Type, src, fset)
		if len(f.Names) == 0 {
			res = append(res, Field{Name: "", Type: typ})
		} else {
			for _, n := range f.Names {
				res = append(res, Field{Name: n.Name, Type: typ})
			}
		}
	}
	return res
}

func hasGoldStandard(doc string, kind string) bool {
    if !strings.Contains(doc, "Summary:") { return false }
    if kind == "func" || kind == "method" {
        if !strings.Contains(doc, "Returns:") { return false }
    }
    return true
}

func main() {
	var nodes []Node
	fset := token.NewFileSet()

	filepath.Walk("server/pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") { return nil }
		if strings.Contains(path, "testutil/") || strings.HasPrefix(filepath.Base(path), "mock_") { return nil }

		src, err := os.ReadFile(path)
		if err != nil { return nil }

		f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil { return nil }

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					doc := ""
					docStart := 0
					docEnd := 0
					if d.Doc != nil {
						doc = d.Doc.Text()
						docStart = fset.Position(d.Doc.Pos()).Line
						docEnd = fset.Position(d.Doc.End()).Line
					}

					kind := "func"
					recv := ""
					if d.Recv != nil {
						kind = "method"
						recv = exprToString(d.Recv.List[0].Type, src, fset)
						if strings.HasPrefix(recv, "*") { recv = recv[1:] }
					}

                    if !hasGoldStandard(doc, kind) {
					    nodes = append(nodes, Node{
						    File:     path,
						    Kind:     kind,
						    Name:     d.Name.Name,
						    Recv:     recv,
						    Params:   getFields(d.Type.Params, src, fset),
						    Results:  getFields(d.Type.Results, src, fset),
						    Doc:      doc,
						    Pos:      fset.Position(d.Pos()).Line,
						    DocStart: docStart,
						    DocEnd:   docEnd,
					    })
                    }
				}
			case *ast.GenDecl:
				if d.Tok == token.TYPE || d.Tok == token.CONST || d.Tok == token.VAR {
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() {
								doc := ""
								docStart := 0
								docEnd := 0
								if d.Doc != nil {
									doc = d.Doc.Text()
									docStart = fset.Position(d.Doc.Pos()).Line
									docEnd = fset.Position(d.Doc.End()).Line
								} else if s.Doc != nil {
									doc = s.Doc.Text()
									docStart = fset.Position(s.Doc.Pos()).Line
									docEnd = fset.Position(s.Doc.End()).Line
								}

                                if !hasGoldStandard(doc, "type") {
								    nodes = append(nodes, Node{
									    File:     path,
									    Kind:     "type",
									    Name:     s.Name.Name,
									    Doc:      doc,
									    Pos:      fset.Position(s.Pos()).Line,
									    DocStart: docStart,
									    DocEnd:   docEnd,
								    })
                                }
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
									doc := ""
									docStart := 0
									docEnd := 0
									if d.Doc != nil {
										doc = d.Doc.Text()
										docStart = fset.Position(d.Doc.Pos()).Line
										docEnd = fset.Position(d.Doc.End()).Line
									} else if s.Doc != nil {
										doc = s.Doc.Text()
										docStart = fset.Position(s.Doc.Pos()).Line
										docEnd = fset.Position(s.Doc.End()).Line
									}

									kind := "const"
									if d.Tok == token.VAR {
										kind = "var"
									}

                                    if !hasGoldStandard(doc, kind) {
									    nodes = append(nodes, Node{
										    File:     path,
										    Kind:     kind,
										    Name:     name.Name,
										    Doc:      doc,
										    Pos:      fset.Position(s.Pos()).Line,
										    DocStart: docStart,
										    DocEnd:   docEnd,
									    })
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

	out, _ := json.Marshal(nodes)
	fmt.Println(string(out))
}
