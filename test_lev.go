package main

import "fmt"

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func main() {
    s1 := "adres"
    s2 := "address"

    n, m := len(s1), len(s2)
    v0 := make([]int, m+1)
    v1 := make([]int, m+1)

    for j := 0; j <= m; j++ {
        v0[j] = j
    }

    for i := 1; i <= n; i++ {
        v1[0] = i
        for j := 1; j <= m; j++ {
            cost := 0
            if s1[i-1] != s2[j-1] {
                cost = 1
            }
            v1[j] = min(v0[j]+1, v1[j-1]+1, v0[j-1]+cost)
        }
        v0, v1 = v1, v0
    }
    fmt.Printf("Standard Lev: %d\n", v0[m])
}
