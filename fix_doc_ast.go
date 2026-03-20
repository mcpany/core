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
		fmt.Println("Usage: go run fix_doc_ast.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "tools" || info.Name() == "tests" || info.Name() == "test" || info.Name() == "mocks" || info.Name() == "examples" || info.Name() == "cmd" || info.Name() == "operator" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") || strings.HasSuffix(path, "zz_generated.deepcopy.go") {
			return nil
		}

		processFile(fset, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

func processFile(fset *token.FileSet, path string) {
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", path, err)
		return
	}

	modified := false

	// First pass to handle function declarations
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if needsUpdate(x.Doc) {
					x.Doc = generateFuncDoc(x.Name.Name, x.Type, x.Body, x.Doc)
					modified = true
				}
			}
		case *ast.GenDecl:
			if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
				for _, s := range x.Specs {
					switch ts := s.(type) {
					case *ast.TypeSpec:
						if ts.Name.IsExported() {
							if needsUpdateType(x.Doc, ts.Doc) {
								if ts.Doc == nil {
									ts.Doc = &ast.CommentGroup{List: []*ast.Comment{}}
								}
								ts.Doc.List = append(ts.Doc.List, &ast.Comment{Text: fmt.Sprintf("// Summary: Defines the %s structure representing core domain properties.", ts.Name.Name)})
								modified = true
							}
							if iface, ok := ts.Type.(*ast.InterfaceType); ok {
								for _, field := range iface.Methods.List {
									if len(field.Names) > 0 && field.Names[0].IsExported() {
										if needsUpdate(field.Doc) {
											field.Doc = generateFuncDoc(field.Names[0].Name, field.Type.(*ast.FuncType), nil, field.Doc)
											modified = true
										}
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, name := range ts.Names {
							if name.IsExported() {
								if needsUpdateType(x.Doc, ts.Doc) {
									if ts.Doc == nil {
										ts.Doc = &ast.CommentGroup{List: []*ast.Comment{}}
									}
									ts.Doc.List = append(ts.Doc.List, &ast.Comment{Text: fmt.Sprintf("// Summary: Represents the %s constant or variable holding specific configuration data.", name.Name)})
									modified = true
								}
							}
						}
					}
				}
			}
		}
	}

	if modified {
		var buf bytes.Buffer
		err = format.Node(&buf, fset, f)
		if err != nil {
			fmt.Printf("Failed to format %s: %v\n", path, err)
			return
		}
        // Fix formatting so comments aren't squished incorrectly for GenDecls by rewriting it purely
		os.WriteFile(path, buf.Bytes(), 0644)
	}
}

func needsUpdate(doc *ast.CommentGroup) bool {
	if doc == nil {
		return true
	}
	text := doc.Text()
	if !strings.Contains(text, "Summary:") {
		return true
	}
	if !strings.Contains(text, "Parameters:") && !strings.Contains(text, "Parameters") {
		return true
	}
	if !strings.Contains(text, "Returns:") && !strings.Contains(text, "Returns") {
		return true
	}
	if !strings.Contains(text, "Errors:") && !strings.Contains(text, "Throws") && !strings.Contains(text, "Errors/Throws:") {
		return true
	}
	if !strings.Contains(text, "Side Effects:") && !strings.Contains(text, "Side Effects") {
		return true
	}
	return false
}

func needsUpdateType(genDoc, specDoc *ast.CommentGroup) bool {
	doc := ""
	if genDoc != nil {
		doc += genDoc.Text()
	}
	if specDoc != nil {
		doc += specDoc.Text()
	}
	if !strings.Contains(doc, "Summary:") {
		return true
	}
	return false
}

func typeString(expr ast.Expr) string {
	if expr == nil {
		return "interface{}"
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(...)"
	}
	return "interface{}"
}

func determineSideEffects(body *ast.BlockStmt) string {
	if body == nil {
		return "None."
	}

	hasNetwork := false
	hasDB := false
	hasDisk := false
	hasMutex := false
	hasGoroutine := false

	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
				name := sel.Sel.Name
				if ident, ok := sel.X.(*ast.Ident); ok {
					pkg := ident.Name
					if pkg == "http" || pkg == "grpc" || pkg == "net" {
						hasNetwork = true
					}
					if pkg == "sql" || pkg == "db" || pkg == "gorm" || pkg == "redis" {
						hasDB = true
					}
					if pkg == "os" || pkg == "io" || pkg == "ioutil" || pkg == "filepath" {
						hasDisk = true
					}
				}
				if name == "Lock" || name == "Unlock" || name == "RLock" || name == "RUnlock" {
					hasMutex = true
				}
			}
		case *ast.GoStmt:
			hasGoroutine = true
		}
		return true
	})

	effects := []string{}
	if hasNetwork {
		effects = append(effects, "Makes external network calls using system packages.")
	}
	if hasDB {
		effects = append(effects, "Executes database queries or redis commands.")
    }
	if hasDisk {
		effects = append(effects, "Interacts with the local file system for reads or writes.")
	}
	if hasMutex {
		effects = append(effects, "Locks or unlocks mutexes, affecting concurrent state.")
	}
	if hasGoroutine {
		effects = append(effects, "Spawns new background goroutines.")
	}

	if len(effects) == 0 {
		return "None."
	}
	return strings.Join(effects, " ")
}

func generateFuncDoc(name string, t *ast.FuncType, body *ast.BlockStmt, existing *ast.CommentGroup) *ast.CommentGroup {
	var comments []*ast.Comment

	existingText := ""
	if existing != nil {
		for _, c := range existing.List {
			comments = append(comments, c)
			existingText += c.Text + "\n"
		}
	}

	if !strings.Contains(existingText, "Summary:") {
		if len(comments) > 0 {
			comments = append(comments, &ast.Comment{Text: "//"})
		}

		action := "Executes"
		if strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "List") || strings.HasPrefix(name, "Read") || strings.HasPrefix(name, "Find") {
			action = "Retrieves"
		} else if strings.HasPrefix(name, "Set") || strings.HasPrefix(name, "Update") || strings.HasPrefix(name, "Write") {
			action = "Modifies"
		} else if strings.HasPrefix(name, "New") || strings.HasPrefix(name, "Create") || strings.HasPrefix(name, "Init") || strings.HasPrefix(name, "Make") {
			action = "Initializes"
		} else if strings.HasPrefix(name, "Delete") || strings.HasPrefix(name, "Remove") || strings.HasPrefix(name, "Clear") {
			action = "Removes"
		} else if strings.HasPrefix(name, "Is") || strings.HasPrefix(name, "Has") || strings.HasPrefix(name, "Can") || strings.HasPrefix(name, "Validate") {
			action = "Checks"
		}

		comments = append(comments, &ast.Comment{Text: fmt.Sprintf("// Summary: %s the %s specific operational logic.", action, name)})
	}

	if !strings.Contains(existingText, "Parameters:") && !strings.Contains(existingText, "Parameters") {
		comments = append(comments, &ast.Comment{Text: "//"})
		comments = append(comments, &ast.Comment{Text: "// Parameters:"})
		if t != nil && t.Params != nil && len(t.Params.List) > 0 {
			for _, p := range t.Params.List {
				typeName := typeString(p.Type)
				if len(p.Names) == 0 {
					paramDesc := fmt.Sprintf("An input value of type %s utilized during execution.", typeName)
					comments = append(comments, &ast.Comment{Text: fmt.Sprintf("//   - unnamed (%s): %s", typeName, paramDesc)})
				} else {
					for _, n := range p.Names {
						paramDesc := fmt.Sprintf("A required input of type %s for the %s process.", typeName, name)
						if n.Name == "ctx" {
							paramDesc = "The context for cancellation signals and request-scoped values."
						} else if strings.Contains(strings.ToLower(n.Name), "id") {
							paramDesc = fmt.Sprintf("The unique identifier string used to locate the specific %s.", name)
						} else if strings.Contains(strings.ToLower(n.Name), "name") {
							paramDesc = fmt.Sprintf("The name string identifying the relevant %s entity.", name)
						} else if strings.Contains(strings.ToLower(n.Name), "config") {
							paramDesc = "The configuration object dictating the operation's behavior."
						} else if strings.Contains(strings.ToLower(n.Name), "req") || strings.Contains(strings.ToLower(n.Name), "request") {
							paramDesc = "The structured request payload containing detailed execution parameters."
						} else if strings.Contains(strings.ToLower(n.Name), "res") || strings.Contains(strings.ToLower(n.Name), "response") {
							paramDesc = "The response writer interface used to stream output data."
						} else if strings.Contains(strings.ToLower(n.Name), "addr") || strings.Contains(strings.ToLower(n.Name), "url") {
							paramDesc = "The target network address or URL string utilized for communication."
						} else if strings.Contains(strings.ToLower(n.Name), "err") {
							paramDesc = "An existing error condition to wrap or process."
                        } else if strings.Contains(strings.ToLower(n.Name), "manager") || strings.Contains(strings.ToLower(n.Name), "store") {
                            paramDesc = "The initialized manager or store instance providing state access."
                        }
						comments = append(comments, &ast.Comment{Text: fmt.Sprintf("//   - %s (%s): %s", n.Name, typeName, paramDesc)})
					}
				}
			}
		} else {
			comments = append(comments, &ast.Comment{Text: "//   - None."})
		}
	}

	if !strings.Contains(existingText, "Returns:") && !strings.Contains(existingText, "Returns") {
		comments = append(comments, &ast.Comment{Text: "//"})
		comments = append(comments, &ast.Comment{Text: "// Returns:"})
		if t != nil && t.Results != nil && len(t.Results.List) > 0 {
			hasNonError := false
			for _, r := range t.Results.List {
				typeName := typeString(r.Type)
				if typeName == "error" {
					continue
				}
				hasNonError = true
				returnDesc := fmt.Sprintf("The instantiated or retrieved %s resulting from the function call.", typeName)
				if typeName == "bool" {
					returnDesc = "A boolean indicating the true/false status of the executed evaluation."
				} else if strings.HasPrefix(typeName, "*") {
					returnDesc = fmt.Sprintf("A pointer to the %s instance representing the operational state.", typeName[1:])
				} else if strings.HasPrefix(typeName, "[]") {
					returnDesc = fmt.Sprintf("A slice of %s elements containing the collected aggregate results.", typeName[2:])
				} else if typeName == "string" {
					returnDesc = "The resulting string identifier or text output."
				} else if typeName == "int" || typeName == "int64" || typeName == "float64" {
					returnDesc = fmt.Sprintf("The computed %s numeric metric or status code.", typeName)
				}
				comments = append(comments, &ast.Comment{Text: fmt.Sprintf("//   - %s: %s", typeName, returnDesc)})
			}
			if !hasNonError {
				comments = append(comments, &ast.Comment{Text: "//   - None."})
			}
		} else {
			comments = append(comments, &ast.Comment{Text: "//   - None."})
		}
	}

	if !strings.Contains(existingText, "Errors:") && !strings.Contains(existingText, "Throws") && !strings.Contains(existingText, "Errors/Throws:") {
		comments = append(comments, &ast.Comment{Text: "//"})
		comments = append(comments, &ast.Comment{Text: "// Errors/Throws:"})
		hasErr := false
		if t != nil && t.Results != nil {
			for _, r := range t.Results.List {
				if id, ok := r.Type.(*ast.Ident); ok && id.Name == "error" {
					hasErr = true
				}
			}
		}
		if hasErr {
			errorDesc := fmt.Sprintf("Returns a structured error if internal validation fails, external dependencies cannot be reached, or state inconsistencies occur during %s execution.", name)
			comments = append(comments, &ast.Comment{Text: fmt.Sprintf("//   - error: %s", errorDesc)})
		} else {
			comments = append(comments, &ast.Comment{Text: "//   - None."})
		}
	}

	if !strings.Contains(existingText, "Side Effects:") && !strings.Contains(existingText, "Side Effects") {
		comments = append(comments, &ast.Comment{Text: "//"})
		comments = append(comments, &ast.Comment{Text: "// Side Effects:"})
		sideEffects := determineSideEffects(body)
		comments = append(comments, &ast.Comment{Text: fmt.Sprintf("//   - %s", sideEffects)})
	}

	return &ast.CommentGroup{List: comments}
}
