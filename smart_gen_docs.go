package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func isPublic(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

func camelToWords(s string) string {
	var words []string
	var currentWord string
	for i, r := range s {
		if unicode.IsUpper(r) {
			if len(currentWord) > 0 {
				words = append(words, currentWord)
			}
			currentWord = string(unicode.ToLower(r))
		} else {
			if i == 0 {
				currentWord = string(unicode.ToLower(r))
			} else {
				currentWord += string(r)
			}
		}
	}
	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}
	return strings.Join(words, " ")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: smart_gen_docs <dir>")
		return
	}
	dir := os.Args[1]

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == "build" || info.Name() == "test" || info.Name() == "tests" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}

		processFile(path)
		return nil
	})
}

type Injection struct {
	Line int
	Text string
}

func processFile(path string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", path, err)
		return
	}

	var injections []Injection

	// Add comments to package
	if f.Doc == nil {
		injections = append(injections, Injection{
			Line: fset.Position(f.Package).Line,
			Text: fmt.Sprintf("// Package %s implements the %s subsystem.", f.Name.Name, f.Name.Name),
		})
	}

	// Iterate over declarations
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.TYPE || d.Tok == token.CONST || d.Tok == token.VAR {
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if isPublic(s.Name.Name) && d.Doc == nil {
							injections = append(injections, Injection{
								Line: fset.Position(d.Pos()).Line,
								Text: fmt.Sprintf("// %s represents the data structure and methods for the %s entity.", s.Name.Name, camelToWords(s.Name.Name)),
							})
						}
					case *ast.ValueSpec:
						if len(s.Names) > 0 && isPublic(s.Names[0].Name) && d.Doc == nil {
							injections = append(injections, Injection{
								Line: fset.Position(d.Pos()).Line,
								Text: fmt.Sprintf("// %s defines a constant or variable for %s.", s.Names[0].Name, camelToWords(s.Names[0].Name)),
							})
						}
					}
				}
			}
		case *ast.FuncDecl:
			if isPublic(d.Name.Name) {
				if d.Doc == nil || !strings.Contains(d.Doc.Text(), "Parameters:") {
					docStr := createFuncDocString(d)
					targetLine := fset.Position(d.Pos()).Line
					if d.Doc != nil {
						targetLine = fset.Position(d.Doc.Pos()).Line
					}
					injections = append(injections, Injection{
						Line: targetLine,
						Text: docStr,
					})
				}
			}
		}
	}

	if len(injections) > 0 {
		applyInjections(path, injections)
	}
}

func getTypeStr(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	} else if sel, ok := expr.(*ast.SelectorExpr); ok {
		if xIdent, ok := sel.X.(*ast.Ident); ok {
			return xIdent.Name + "." + sel.Sel.Name
		}
	} else if star, ok := expr.(*ast.StarExpr); ok {
		return "*" + getTypeStr(star.X)
	} else if arr, ok := expr.(*ast.ArrayType); ok {
		return "[]" + getTypeStr(arr.Elt)
	} else if mapT, ok := expr.(*ast.MapType); ok {
		return "map[" + getTypeStr(mapT.Key) + "]" + getTypeStr(mapT.Value)
	} else if _, ok := expr.(*ast.InterfaceType); ok {
		return "interface{}"
	} else if _, ok := expr.(*ast.FuncType); ok {
		return "func()"
	} else if ell, ok := expr.(*ast.Ellipsis); ok {
		return "..." + getTypeStr(ell.Elt)
	} else if chanT, ok := expr.(*ast.ChanType); ok {
		return "chan " + getTypeStr(chanT.Value)
	}
	return "unknown"
}

func createFuncDocString(d *ast.FuncDecl) string {
	var comments []string
	name := d.Name.Name

	action := "handles"
	target := camelToWords(name)
	if strings.HasPrefix(name, "Get") {
		action = "retrieves"
		target = "the " + camelToWords(strings.TrimPrefix(name, "Get"))
	} else if strings.HasPrefix(name, "Set") {
		action = "updates"
		target = "the " + camelToWords(strings.TrimPrefix(name, "Set"))
	} else if strings.HasPrefix(name, "New") {
		action = "initializes and returns a new"
		target = camelToWords(strings.TrimPrefix(name, "New")) + " instance"
	} else if strings.HasPrefix(name, "List") {
		action = "returns a collection of"
		target = camelToWords(strings.TrimPrefix(name, "List"))
	} else if strings.HasPrefix(name, "Update") {
		action = "modifies an existing"
		target = camelToWords(strings.TrimPrefix(name, "Update"))
	} else if strings.HasPrefix(name, "Delete") {
		action = "removes the specified"
		target = camelToWords(strings.TrimPrefix(name, "Delete")) + " from the system"
	} else if strings.HasPrefix(name, "Create") {
		action = "provisions a new"
		target = camelToWords(strings.TrimPrefix(name, "Create"))
	} else if strings.HasPrefix(name, "Is") || strings.HasPrefix(name, "Has") {
		action = "checks if"
		target = "the condition " + camelToWords(name) + " is met"
	}

	comments = append(comments, fmt.Sprintf("// %s %s %s.", name, action, target))
	comments = append(comments, "//")
	comments = append(comments, "// Parameters:")

	hasParams := false
	if d.Type.Params != nil && len(d.Type.Params.List) > 0 {
		hasParams = true
		for _, field := range d.Type.Params.List {
			typeStr := getTypeStr(field.Type)
			if len(field.Names) > 0 {
				for _, n := range field.Names {
					if n.Name == "_" {
						comments = append(comments, fmt.Sprintf("//   - %s (%s): Reserved parameter, currently unused.", n.Name, typeStr))
					} else if typeStr == "context.Context" {
						comments = append(comments, fmt.Sprintf("//   - %s (%s): The execution context for managing deadlines and cancellations.", n.Name, typeStr))
					} else if strings.Contains(strings.ToLower(n.Name), "id") {
						comments = append(comments, fmt.Sprintf("//   - %s (%s): The unique identifier for the %s.", n.Name, typeStr, n.Name))
					} else if strings.Contains(strings.ToLower(n.Name), "req") || strings.Contains(strings.ToLower(n.Name), "request") {
						comments = append(comments, fmt.Sprintf("//   - %s (%s): The request payload containing necessary arguments.", n.Name, typeStr))
					} else {
						comments = append(comments, fmt.Sprintf("//   - %s (%s): The %s input required for processing.", n.Name, typeStr, n.Name))
					}
				}
			} else {
				comments = append(comments, fmt.Sprintf("//   - unnamed (%s): An unnamed argument.", typeStr))
			}
		}
	}
	if !hasParams {
		comments = append(comments, "//   - None")
	}
	comments = append(comments, "//")
	comments = append(comments, "// Returns:")

	if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		for _, field := range d.Type.Results.List {
			typeStr := getTypeStr(field.Type)
			if typeStr == "error" {
				comments = append(comments, "//   - error: Returns an error if the execution fails or validation does not pass.")
			} else if strings.HasPrefix(typeStr, "[]") {
				comments = append(comments, fmt.Sprintf("//   - %s: A slice containing the requested elements.", typeStr))
			} else if typeStr == "bool" {
				comments = append(comments, fmt.Sprintf("//   - %s: A boolean indicating the result of the condition.", typeStr))
			} else {
				comments = append(comments, fmt.Sprintf("//   - %s: The generated or retrieved entity.", typeStr))
			}
		}
	} else {
		comments = append(comments, "//   - None")
	}
	comments = append(comments, "//")
	comments = append(comments, "// Errors:")

	hasError := false
	if d.Type.Results != nil {
		for _, field := range d.Type.Results.List {
			typeStr := getTypeStr(field.Type)
			if typeStr == "error" {
				hasError = true
			}
		}
	}

	if hasError {
		comments = append(comments, "//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.")
	} else {
		comments = append(comments, "//   - None.")
	}
	comments = append(comments, "//")
	comments = append(comments, "// Side Effects:")

	sideEffects := "None."
	if strings.HasPrefix(name, "Update") || strings.HasPrefix(name, "Set") || strings.HasPrefix(name, "Create") || strings.HasPrefix(name, "Delete") {
		sideEffects = "Modifies internal state or updates external storage components."
	} else if strings.Contains(strings.ToLower(name), "fetch") || strings.Contains(strings.ToLower(name), "load") {
		sideEffects = "May trigger network calls or perform disk I/O operations."
	}
	comments = append(comments, fmt.Sprintf("//   - %s", sideEffects))

	return strings.Join(comments, "\n")
}

func applyInjections(path string, injections []Injection) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")

	// Sort injections descending so line numbers don't shift as we insert
	sort.Slice(injections, func(i, j int) bool {
		return injections[i].Line > injections[j].Line
	})

	for _, inj := range injections {
		insertIdx := inj.Line - 1
		if insertIdx < 0 { insertIdx = 0 }

		// Remove existing comments that are adjacent above insertIdx
		for insertIdx > 0 && strings.HasPrefix(strings.TrimSpace(lines[insertIdx-1]), "//") {
			// Delete the comment line
			lines = append(lines[:insertIdx-1], lines[insertIdx:]...)
			insertIdx--
		}

		// Insert the new text
		newLines := strings.Split(inj.Text, "\n")

        // Let's copy properly
        var newContent []string
        newContent = append(newContent, lines[:insertIdx]...)
        newContent = append(newContent, newLines...)
        newContent = append(newContent, lines[insertIdx:]...)
        lines = newContent
	}

	os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
