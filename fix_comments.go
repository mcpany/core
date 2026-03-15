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
	err := filepath.Walk("server", func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/vendor/") || strings.Contains(path, "proto/") || strings.Contains(path, "build/") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil { return nil }

		changed := false

		for _, decl := range node.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							doc := s.Doc
							if doc == nil && len(d.Specs) == 1 {
								doc = d.Doc
							}

							hasDoc := doc != nil
							var docText string
							if hasDoc { docText = doc.Text() }

							if !hasDoc || !strings.Contains(docText, "Summary:") {
								text := strings.TrimSpace(docText)
								text = strings.ReplaceAll(text, "\n", " ")
								if text == "" {
									text = fmt.Sprintf("The %s type.", s.Name.Name)
								}
								newDoc := fmt.Sprintf("%s ...\n\nSummary: %s", s.Name.Name, text)

								cg := &ast.CommentGroup{
									List: []*ast.Comment{
										{Text: "/*\n" + newDoc + "\n*/"},
									},
								}

								if len(d.Specs) == 1 {
									d.Doc = cg
								} else {
									s.Doc = cg
								}
								changed = true
							}
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.IsExported() {
								doc := s.Doc
								if doc == nil && len(d.Specs) == 1 {
									doc = d.Doc
								}

								hasDoc := doc != nil
								var docText string
								if hasDoc { docText = doc.Text() }

								if !hasDoc || !strings.Contains(docText, "Summary:") {
									text := strings.TrimSpace(docText)
									text = strings.ReplaceAll(text, "\n", " ")
									if text == "" {
										text = fmt.Sprintf("The %s value.", name.Name)
									}
									newDoc := fmt.Sprintf("%s ...\n\nSummary: %s", name.Name, text)

									cg := &ast.CommentGroup{
										List: []*ast.Comment{
											{Text: "/*\n" + newDoc + "\n*/"},
										},
									}
									if len(d.Specs) == 1 {
										d.Doc = cg
									} else {
										s.Doc = cg
									}
									changed = true
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					doc := d.Doc
					hasDoc := doc != nil
					var docText string
					if hasDoc { docText = doc.Text() }

					if !hasDoc || !strings.Contains(docText, "Summary:") || !strings.Contains(docText, "Parameters:") || !strings.Contains(docText, "Returns:") || !strings.Contains(docText, "Errors:") || !strings.Contains(docText, "Side Effects:") {

						summary := fmt.Sprintf("%s ...", d.Name.Name)
						if docText != "" {
							lines := strings.Split(docText, "\n")
							summary = strings.TrimSpace(lines[0])
						}

						var params []string
						if d.Type.Params != nil {
							for _, p := range d.Type.Params.List {
								var buf bytes.Buffer
								format.Node(&buf, fset, p.Type)
								typ := buf.String()

								if len(p.Names) > 0 {
									for _, n := range p.Names {
										params = append(params, fmt.Sprintf("  - %s (%s): The %s parameter.", n.Name, typ, n.Name))
									}
								} else {
									params = append(params, fmt.Sprintf("  - (%s): The parameter.", typ))
								}
							}
						}

						var returns []string
						if d.Type.Results != nil {
							for _, r := range d.Type.Results.List {
								var buf bytes.Buffer
								format.Node(&buf, fset, r.Type)
								typ := buf.String()

								if len(r.Names) > 0 {
									for _, n := range r.Names {
										returns = append(returns, fmt.Sprintf("  - %s: The resulting %s.", typ, n.Name))
									}
								} else {
									if typ == "error" {
										returns = append(returns, fmt.Sprintf("  - error: An error if the operation fails."))
									} else {
										returns = append(returns, fmt.Sprintf("  - %s: The resulting %s.", typ, typ))
									}
								}
							}
						}

						var newDoc string
						newDoc += fmt.Sprintf("%s\n\nSummary: %s\n\nParameters:\n", d.Name.Name, summary)
						if len(params) == 0 {
							newDoc += "  - None\n"
						} else {
							newDoc += strings.Join(params, "\n") + "\n"
						}
						newDoc += "\nReturns:\n"
						if len(returns) == 0 {
							newDoc += "  - None\n"
						} else {
							newDoc += strings.Join(returns, "\n") + "\n"
						}
						newDoc += "\nErrors:\n  - Returns an error if the operation fails or is invalid.\n\nSide Effects:\n  - None"

						d.Doc = &ast.CommentGroup{
							List: []*ast.Comment{
								{Text: "/*\n" + newDoc + "\n*/"},
							},
						}
						changed = true
					}
				}
			}
		}

		if changed {
			var buf bytes.Buffer
			if err := format.Node(&buf, fset, node); err != nil {
				return err
			}
			if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
				return err
			}
			fmt.Printf("Updated %s\n", path)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}
