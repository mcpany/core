package main

import (
	"fmt"
	"strings"
	"unicode"
)

func isWordChar(r byte) bool {
	return unicode.IsLetter(rune(r)) || unicode.IsDigit(rune(r)) || r == '_'
}

func analyzeQuoteContext(val string, idx int) bool {
	singleQuote := 0
	doubleQuote := 0
	for i := 0; i < idx; i++ {
		if val[i] == '\'' {
			singleQuote++
		} else if val[i] == '"' {
			doubleQuote++
		}
	}
	return singleQuote%2 != 0 || doubleQuote%2 != 0
}

func checkSQLKeywords(val string) error {
	upperVal := strings.ToUpper(val)
	keywords := []string{
		"OR", "AND", "UNION", "SELECT", "FROM", "WHERE", "JOIN",
		"DROP", "ALTER", "CREATE", "INSERT", "UPDATE", "DELETE",
		"--",
	}

	for _, kwI := range keywords {
		kw := strings.ToUpper(kwI)
		if kw == "--" {
			if strings.Contains(upperVal, "--") {
				return fmt.Errorf("SQL injection detected: value contains '--'")
			}
			continue
		}

		idx := strings.Index(upperVal, kw)
		for {
			if idx == -1 {
				break
			}
			// check if it's quoted
			if analyzeQuoteContext(upperVal, idx) {
				// it's quoted, safe to skip
				upperVal = upperVal[idx+len(kw):]
				continue
			}

			// verify it's a separate word, not part of another word like "FORMAT" containing "OR"
			startOk := idx == 0 || !unicode.IsLetter(rune(upperVal[idx-1]))
			endOk := idx+len(kw) == len(upperVal) || !unicode.IsLetter(rune(upperVal[idx+len(kw)]))
			if startOk && endOk {
				return fmt.Errorf("SQL injection detected: value contains SQL keyword %q in unquoted context", kw)
			}
			nextIdx := strings.Index(upperVal[idx+len(kw):], kw)
			if nextIdx == -1 {
				break
			}
			idx += len(kw) + nextIdx
		}
	}
	return nil
}

func main() {
	tests := []string{
		"val OR 1=1",
		"FORMAT OR 1=1",
		"1; DROP TABLE users; --",
		"1; SELECT 1",
	}
	for _, val := range tests {
		fmt.Printf("testing %q: %v\n", val, checkSQLKeywords(val) != nil)
	}
}
