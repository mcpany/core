package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("bazelisk", "run", "//:lint")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))
}
