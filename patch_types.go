package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	b, err := ioutil.ReadFile("server/pkg/tool/types.go")
	if err != nil {
		panic(err)
	}

	content := string(b)

	search_fn := `//nolint:gocyclo
func checkUnquotedKeywords(val string, keywords []string) error {`

	replace_fn := `//nolint:gocyclo
func checkUnquotedKeywords(val string, keywords []string) error {
	for _, kw := range keywords {
		idx := strings.Index(val, kw)
		if idx != -1 {
			leftOk := idx == 0 || !isWordChar(val[idx-1])
			rightOk := idx+len(kw) == len(val) || !isWordChar(val[idx+len(kw)])
			if leftOk && rightOk {
				if idx > 0 {
					lastChar := val[idx-1]
					if lastChar == '$' || lastChar == '@' || lastChar == '%' || lastChar == '>' {
						continue
					}
				}

				return fmt.Errorf("interpreter injection detected: dangerous keyword %q found (unquoted)", kw)
			}
		}
	}
	return nil
}

func checkUnquotedKeywords_old(val string, keywords []string) error {`

	if strings.Contains(content, search_fn) {
		content = strings.Replace(content, search_fn, replace_fn, 1)
		ioutil.WriteFile("server/pkg/tool/types.go", []byte(content), 0644)
		fmt.Println("Patched successfully")
	} else {
		fmt.Println("Could not find function")
	}
}
