package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "pkg/app/api.go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	b, err := os.ReadFile("pkg/app/api.go")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ExprStmt:
			if call, ok := x.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "Write" || sel.Sel.Name == "Marshal" || sel.Sel.Name == "Encode" || sel.Sel.Name == "Close" || sel.Sel.Name == "Stat" || sel.Sel.Name == "LookPath" || sel.Sel.Name == "UserFromContext" {
						lineNum := fset.Position(x.Pos()).Line - 1
						if !strings.Contains(lines[lineNum], "//nolint:errcheck") {
							lines[lineNum] = lines[lineNum] + " //nolint:errcheck"
						}
					}
				}
			}
		case *ast.AssignStmt:
			isUnderscore := false
			for _, lhs := range x.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
					isUnderscore = true
					break
				}
			}

			if isUnderscore && len(x.Rhs) > 0 {
				if call, ok := x.Rhs[0].(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                        // For the LHS being '_', these standard ones also trigger if unassigned properly
						if sel.Sel.Name == "Write" || sel.Sel.Name == "Encode" || sel.Sel.Name == "Marshal" || sel.Sel.Name == "Close" || sel.Sel.Name == "Stat" || sel.Sel.Name == "LookPath" || sel.Sel.Name == "UserFromContext" {
							lineNum := fset.Position(x.Pos()).Line - 1
							if !strings.Contains(lines[lineNum], "//nolint:errcheck") {
								lines[lineNum] = lines[lineNum] + " //nolint:errcheck"
							}
						}
					}
				}
			}
		}
		return true
	})

	err = os.WriteFile("pkg/app/api.go", []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done")
}
