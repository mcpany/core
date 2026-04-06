package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// niceName split camelCase to readable words
func niceName(s string) string {
	if len(s) == 0 {
		return ""
	}
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	s2 := re.ReplaceAllString(s, "${1} ${2}")
	return strings.ToLower(s2)
}

func getDesc(name string, typeName string) string {
	if name == "ctx" {
		return "The request context."
	} else if name == "err" {
		return "The error."
	} else if name == "req" {
		return "The request object."
	} else if name == "res" {
		return "The response object."
	} else if name == "_" {
		return "Unused parameter."
	}

	desc := "The " + niceName(name) + "."
	if strings.Contains(strings.ToLower(desc), "error") || typeName == "error" {
		return "An error if the operation fails."
	}
	return desc
}

func exprToString(expr ast.Expr, src []byte) string {
	if expr == nil {
		return ""
	}
	return string(src[expr.Pos()-1 : expr.End()-1])
}

func processFile(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	src, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	lines := bytes.Split(src, []byte("\n"))
	modified := false

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !fn.Name.IsExported() || fn.Doc == nil {
			continue
		}

		params := []string{}
		if fn.Type.Params != nil {
			for _, field := range fn.Type.Params.List {
				typeStr := exprToString(field.Type, src)
				if len(field.Names) == 0 {
					// Unnamed param
					desc := "The parameter."
					if typeStr == "error" {
						desc = getDesc("", typeStr)
					}
					params = append(params, fmt.Sprintf("//   - %s: %s", typeStr, desc))
				} else {
					for _, name := range field.Names {
						params = append(params, fmt.Sprintf("//   - %s (%s): %s", name.Name, typeStr, getDesc(name.Name, typeStr)))
					}
				}
			}
		}

		returns := []string{}
		hasErrorRet := false
		if fn.Type.Results != nil {
			for _, field := range fn.Type.Results.List {
				typeStr := exprToString(field.Type, src)
				if len(field.Names) == 0 {
					desc := "The result."
					if typeStr == "error" {
						desc = getDesc("", typeStr)
						hasErrorRet = true
					}
					returns = append(returns, fmt.Sprintf("//   - %s: %s", typeStr, desc))
				} else {
					for _, name := range field.Names {
						desc := getDesc(name.Name, typeStr)
						if typeStr == "error" {
							hasErrorRet = true
						}
						returns = append(returns, fmt.Sprintf("//   - %s: %s", typeStr, desc))
					}
				}
			}
		}

		for _, comment := range fn.Doc.List {
			lineIdx := fset.Position(comment.Pos()).Line - 1
			line := string(lines[lineIdx])

			if strings.Contains(line, "TODO: Document parameters.") {
				newLines := ""
				if len(params) > 0 {
					newLines = strings.Join(params, "\n")
				} else {
					newLines = "//   - None."
				}
				lines[lineIdx] = []byte(newLines)
				modified = true
			} else if strings.Contains(line, "TODO: Document returns.") {
				newLines := ""
				if len(returns) > 0 {
					newLines = strings.Join(returns, "\n")
				} else {
					newLines = "//   - None."
				}
				lines[lineIdx] = []byte(newLines)
				modified = true
			} else if strings.Contains(line, "TODO: Document errors.") {
				if hasErrorRet {
					lines[lineIdx] = []byte("//   - Returns an error if the operation fails.")
				} else {
					lines[lineIdx] = []byte("//   - None.")
				}
				modified = true
			}
		}
	}

	if modified {
		return ioutil.WriteFile(path, bytes.Join(lines, []byte("\n")), 0644)
	}

	return nil
}

func main() {
	err := filepath.Walk("server", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.Contains(path, "vendor") && !strings.Contains(path, "node_modules") {
			processFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
