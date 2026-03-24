package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Add a helper to convert CamelCase to separate words
func camelToWords(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	return strings.ToLower(re.ReplaceAllString(s, "${1} ${2}"))
}

func getTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return getTypeName(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeName(e.X)
	case *ast.ArrayType:
		return "[]" + getTypeName(e.Elt)
	case *ast.MapType:
		return "map[" + getTypeName(e.Key) + "]" + getTypeName(e.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan"
	case *ast.Ellipsis:
		return "..." + getTypeName(e.Elt)
	case *ast.StructType:
		return "struct"
	default:
		return fmt.Sprintf("%T", e)
	}
}

type insert struct {
	offset int
	text   string
}

func processFile(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(contentBytes)

	var inserts []insert

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				hasDoc := d.Doc != nil
				docText := ""
				if hasDoc {
					docText = d.Doc.Text()
				}

				var newDoc []string

				// Generate Summary
				if !hasDoc {
					action := "Executes"
					words := camelToWords(d.Name.Name)
					if strings.HasPrefix(words, "get ") {
						action = "Retrieves"
					} else if strings.HasPrefix(words, "set ") {
						action = "Updates"
					} else if strings.HasPrefix(words, "is ") || strings.HasPrefix(words, "has ") {
						action = "Checks if"
					} else if strings.HasPrefix(words, "new ") {
						action = "Initializes"
					}
					newDoc = append(newDoc, fmt.Sprintf("// %s %s %s.", d.Name.Name, strings.ToLower(action), words))
				}
				if !strings.Contains(docText, "Summary:") {
					newDoc = append(newDoc, "//")
					action := "Performs"
					words := camelToWords(d.Name.Name)
					if strings.HasPrefix(words, "get ") { action = "Retrieves" }
					if strings.HasPrefix(words, "set ") { action = "Updates" }
					if strings.HasPrefix(words, "new ") { action = "Initializes" }
					newDoc = append(newDoc, fmt.Sprintf("// Summary: %s the %s operation.", action, words))
				}

				// Generate Parameters
				if !strings.Contains(docText, "Parameters:") {
					newDoc = append(newDoc, "//")
					newDoc = append(newDoc, "// Parameters:")
					if d.Type.Params != nil && len(d.Type.Params.List) > 0 {
						for _, field := range d.Type.Params.List {
							typeName := getTypeName(field.Type)
							names := field.Names
							if len(names) == 0 {
								newDoc = append(newDoc, fmt.Sprintf("//   - unnamed (%s): The %s parameter.", typeName, typeName))
							} else {
								for _, name := range names {
									newDoc = append(newDoc, fmt.Sprintf("//   - %s (%s): The %s argument.", name.Name, typeName, name.Name))
								}
							}
						}
					} else {
						newDoc = append(newDoc, "//   - None")
					}
				}

				// Generate Returns
				hasErrorReturn := false
				if !strings.Contains(docText, "Returns:") {
					newDoc = append(newDoc, "//")
					newDoc = append(newDoc, "// Returns:")
					if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
						for _, field := range d.Type.Results.List {
							typeName := getTypeName(field.Type)
							if typeName == "error" { hasErrorReturn = true }

							names := field.Names
							if len(names) == 0 {
								desc := "The result of the operation"
								if typeName == "error" { desc = "An error if the operation fails" }
								newDoc = append(newDoc, fmt.Sprintf("//   - %s: %s.", typeName, desc))
							} else {
								for _, name := range names {
									desc := "The returned " + name.Name
									if typeName == "error" { desc = "An error if the operation fails" }
									newDoc = append(newDoc, fmt.Sprintf("//   - %s (%s): %s.", name.Name, typeName, desc))
								}
							}
						}
					} else {
						newDoc = append(newDoc, "//   - None")
					}
				} else {
					if d.Type.Results != nil {
						for _, field := range d.Type.Results.List {
							if getTypeName(field.Type) == "error" { hasErrorReturn = true }
						}
					}
				}

				// Generate Errors
				if !strings.Contains(docText, "Errors:") {
					newDoc = append(newDoc, "//")
					newDoc = append(newDoc, "// Errors:")
					if hasErrorReturn {
						newDoc = append(newDoc, "//   - Returns an error if the operation encounters an issue.")
					} else {
						newDoc = append(newDoc, "//   - None")
					}
				}

				// Generate Side Effects
				if !strings.Contains(docText, "Side Effects:") {
					newDoc = append(newDoc, "//")
					newDoc = append(newDoc, "// Side Effects:")
					// Simple heuristic
					if strings.Contains(strings.ToLower(d.Name.Name), "save") || strings.Contains(strings.ToLower(d.Name.Name), "write") || strings.Contains(strings.ToLower(d.Name.Name), "update") || strings.Contains(strings.ToLower(d.Name.Name), "delete") {
						newDoc = append(newDoc, "//   - Modifies internal state or writes to external storage.")
					} else {
						newDoc = append(newDoc, "//   - None")
					}
				}

				if len(newDoc) > 0 {
					if hasDoc {
						lastComment := d.Doc.List[len(d.Doc.List)-1]
						offset := fset.Position(lastComment.End()).Offset
						inserts = append(inserts, insert{offset: offset, text: "\n" + strings.Join(newDoc, "\n")})
					} else {
						offset := fset.Position(d.Pos()).Offset
						inserts = append(inserts, insert{offset: offset, text: strings.Join(newDoc, "\n") + "\n"})
					}
				}
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						hasDoc := d.Doc != nil || s.Doc != nil
						docText := ""
						if d.Doc != nil { docText += d.Doc.Text() }
						if s.Doc != nil { docText += s.Doc.Text() }

						var newDoc []string
						if !hasDoc {
							words := camelToWords(s.Name.Name)
							newDoc = append(newDoc, fmt.Sprintf("// %s defines the %s structure.", s.Name.Name, words))
						}
						if !strings.Contains(docText, "Summary:") {
							newDoc = append(newDoc, "//")
							newDoc = append(newDoc, fmt.Sprintf("// Summary: Represents the %s entity.", s.Name.Name))
						}

						if len(newDoc) > 0 {
							if hasDoc {
								var lastComment *ast.Comment
								if s.Doc != nil { lastComment = s.Doc.List[len(s.Doc.List)-1] } else { lastComment = d.Doc.List[len(d.Doc.List)-1] }
								offset := fset.Position(lastComment.End()).Offset
								inserts = append(inserts, insert{offset: offset, text: "\n" + strings.Join(newDoc, "\n")})
							} else {
								offset := fset.Position(d.Pos()).Offset
								inserts = append(inserts, insert{offset: offset, text: strings.Join(newDoc, "\n") + "\n"})
							}
						}
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							hasDoc := d.Doc != nil || s.Doc != nil
							docText := ""
							if d.Doc != nil { docText += d.Doc.Text() }
							if s.Doc != nil { docText += s.Doc.Text() }

							var newDoc []string
							if !hasDoc {
								newDoc = append(newDoc, fmt.Sprintf("// %s specifies the %s value.", name.Name, camelToWords(name.Name)))
							}
							if !strings.Contains(docText, "Summary:") {
								newDoc = append(newDoc, "//")
								newDoc = append(newDoc, fmt.Sprintf("// Summary: Defines the %s constant or variable.", name.Name))
							}

							if len(newDoc) > 0 {
								if hasDoc {
									var lastComment *ast.Comment
									if s.Doc != nil { lastComment = s.Doc.List[len(s.Doc.List)-1] } else { lastComment = d.Doc.List[len(d.Doc.List)-1] }
									offset := fset.Position(lastComment.End()).Offset
									inserts = append(inserts, insert{offset: offset, text: "\n" + strings.Join(newDoc, "\n")})
								} else {
									offset := fset.Position(d.Pos()).Offset
									inserts = append(inserts, insert{offset: offset, text: strings.Join(newDoc, "\n") + "\n"})
								}
							}
						}
					}
				}
			}
		}
	}

	if len(inserts) > 0 {
		// sort inserts by offset descending
		for i := 0; i < len(inserts); i++ {
			for j := i + 1; j < len(inserts); j++ {
				if inserts[i].offset < inserts[j].offset {
					inserts[i], inserts[j] = inserts[j], inserts[i]
				}
			}
		}

		var out string = content
		for _, ins := range inserts {
			out = out[:ins.offset] + ins.text + out[ins.offset:]
		}

		return os.WriteFile(path, []byte(out), 0644)
	}

	return nil
}

func main() {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "vendor") || strings.Contains(path, "proto") || strings.Contains(path, "zz_generated") {
			return nil
		}

		if !strings.HasPrefix(path, "server/") && !strings.HasPrefix(path, "k8s/") {
			return nil
		}

		err = processFile(path)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
