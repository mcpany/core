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

type SymbolInfo struct {
	File       string `json:"file"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	DocPos     int    `json:"doc_pos"`
	DocStart   int    `json:"doc_start"`
	DocEnd     int    `json:"doc_end"`
	Params     []string `json:"params"`
	Returns    []string `json:"returns"`
	HasError   bool `json:"has_error"`
	Existing   string `json:"existing"`
}

func main() {
	fset := token.NewFileSet()
	var symbols []SymbolInfo

	dirs := []string{"server", "src"}
	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil { return err }
			if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if err != nil { return err }

				tf := fset.File(node.Pos())
				content, _ := os.ReadFile(path)

				for _, decl := range node.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						if d.Name.IsExported() {
							doc := ""
							docStart, docEnd := -1, -1
							if d.Doc != nil {
								doc = d.Doc.Text()
								docStart = tf.Offset(d.Doc.Pos())
								docEnd = tf.Offset(d.Doc.End())
							}

							if !strings.Contains(doc, "Summary:") || !strings.Contains(doc, "Returns:") || (!strings.Contains(doc, "Errors:") && !strings.Contains(doc, "Throws/Errors:")) {
								var params []string
								if d.Type.Params != nil {
									for _, p := range d.Type.Params.List {
										typeStr := getTypeStr(p.Type, content, tf)
										if len(p.Names) == 0 {
											params = append(params, fmt.Sprintf("_|%s", typeStr))
										} else {
											for _, n := range p.Names {
												params = append(params, fmt.Sprintf("%s|%s", n.Name, typeStr))
											}
										}
									}
								}

								var returns []string
								hasError := false
								if d.Type.Results != nil {
									for _, r := range d.Type.Results.List {
										typeStr := getTypeStr(r.Type, content, tf)
										if typeStr == "error" {
											hasError = true
										}
										if len(r.Names) == 0 {
											returns = append(returns, typeStr)
										} else {
											for range r.Names {
												returns = append(returns, typeStr)
											}
										}
									}
								}

								insertPos := tf.Offset(d.Pos())
								symbols = append(symbols, SymbolInfo{
									File: path, Kind: "Func", Name: d.Name.Name,
									DocPos: insertPos, DocStart: docStart, DocEnd: docEnd,
									Params: params, Returns: returns, HasError: hasError, Existing: doc,
								})
							}
						}
					case *ast.GenDecl:
						for _, spec := range d.Specs {
							switch s := spec.(type) {
							case *ast.TypeSpec:
								if s.Name.IsExported() {
									doc := ""
									docStart, docEnd := -1, -1
									if d.Doc != nil {
										doc = d.Doc.Text()
										docStart = tf.Offset(d.Doc.Pos())
										docEnd = tf.Offset(d.Doc.End())
									}
									if s.Doc != nil {
										doc = s.Doc.Text()
										docStart = tf.Offset(s.Doc.Pos())
										docEnd = tf.Offset(s.Doc.End())
									}
									if !strings.Contains(doc, "Summary:") || !strings.Contains(doc, "Returns:") || (!strings.Contains(doc, "Errors:") && !strings.Contains(doc, "Throws/Errors:")) {
										insertPos := tf.Offset(d.Pos())
										if d.Lparen.IsValid() {
											insertPos = tf.Offset(s.Pos())
										}
										symbols = append(symbols, SymbolInfo{
											File: path, Kind: "Type", Name: s.Name.Name,
											DocPos: insertPos, DocStart: docStart, DocEnd: docEnd,
											Existing: doc,
										})
									}
								}
							case *ast.ValueSpec:
								for _, name := range s.Names {
									if name.IsExported() {
										doc := ""
										docStart, docEnd := -1, -1
										if d.Doc != nil {
											doc = d.Doc.Text()
											docStart = tf.Offset(d.Doc.Pos())
											docEnd = tf.Offset(d.Doc.End())
										}
										if s.Doc != nil {
											doc = s.Doc.Text()
											docStart = tf.Offset(s.Doc.Pos())
											docEnd = tf.Offset(s.Doc.End())
										}
										if !strings.Contains(doc, "Summary:") || !strings.Contains(doc, "Returns:") || (!strings.Contains(doc, "Errors:") && !strings.Contains(doc, "Throws/Errors:")) {
											insertPos := tf.Offset(d.Pos())
											if d.Lparen.IsValid() {
												insertPos = tf.Offset(s.Pos())
											}
											symbols = append(symbols, SymbolInfo{
												File: path, Kind: "Var", Name: name.Name,
												DocPos: insertPos, DocStart: docStart, DocEnd: docEnd,
												Existing: doc,
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
	}

	b, _ := json.MarshalIndent(symbols, "", "  ")
	os.WriteFile("symbols.json", b, 0644)
	fmt.Printf("Extracted %d symbols to symbols.json\n", len(symbols))
}

func getTypeStr(expr ast.Expr, content []byte, tf *token.File) string {
	if expr == nil { return "" }
	start := tf.Offset(expr.Pos())
	end := tf.Offset(expr.End())
	if start >= 0 && end <= len(content) {
		return string(content[start:end])
	}
	return "unknown"
}
