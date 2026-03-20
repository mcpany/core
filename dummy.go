package main

import "fmt"

func main() {
    ring := make([]int, 0, 5)
    for i := 0; i < 5; i++ {
        ring = append(ring, i)
    }

    // ring = [0, 1, 2, 3, 4]
    // append(ring[3:], ring[:3]...)
    // ring[3:] = [3, 4] (len=2, cap=2)
    // cap = 5 - 3 = 2! Wait!
    fmt.Printf("cap of ring[3:]: %d\n", cap(ring[3:]))

    res := append(ring[3:], ring[:3]...)
    fmt.Printf("res: %v\n", res)
    fmt.Printf("ring after: %v\n", ring)
}
