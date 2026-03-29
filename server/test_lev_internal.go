package main

import (
    "fmt"
    "github.com/mcpany/core/server/pkg/util"
)

func main() {
    d1 := util.LevenshteinDistance("adres", "address")
    d2 := util.LevenshteinDistance("adres", "args")
    fmt.Printf("adres vs address: %d\n", d1)
    fmt.Printf("adres vs args: %d\n", d2)
}
