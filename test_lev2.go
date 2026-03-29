package main
import "fmt"
func main() {
    v0 := []int{0, 1, 2, 3, 4}
    v1 := []int{1, 1, 2, 3, 4}
    // we want v1 values into v0, then we want v1 to be a new slice (not really new, we just overwrite it next iteration)
    // the problem is if we do `v0, v1 = v1, v0` we swap pointers. But if they are slices of the SAME stack array they might overlap if not sliced properly, but they are sliced disjointly:
    // v0 = stackBuf[:m+1]
    // v1 = stackBuf[m+1:2*(m+1)]
    // Swapping `v0, v1 = v1, v0` is 100% fine in Go, even if they point to the same underlying array, because they are disjoint slices.
    // Let's print out what actually happens.

    // The bug is that `v0[k] = v1[k]` copies values. BUT then `v1[0] = i` in the NEXT iteration overwrites the `v1` which is STILL the SAME slice!
    // So both v0 and v1 end up with the same values, AND THEN we overwrite `v1[j] = min(v0[j]+1, ...)` meaning `v0` is mutating while we read it because they are identical? NO! `v0` and `v1` are different slices. If we just copy `v0[k] = v1[k]`, then `v0` has the old `v1` values. But what about the next row? `v1` gets overwritten. BUT `v1` was used to store the PREVIOUS row!
    // Oh, if we just do `for k := 0; k <= m; k++ { v0[k] = v1[k] }`, then we copy `v1` values into `v0`. In the next iteration, `v1` is overwritten with NEW values. `v0` remains the OLD `v1` values. This is logically correct!
    // But `adres` vs `args` returns `2` instead of what it should be? What should it be?
    // a r g s
    // a d r e s
    // a(0) -> a(0). d(1) vs r(1). r(2) vs g(2). e(3) vs s(3). s(4).
    // edit distance between `adres` and `address` is 2. `adres` and `args` is 3.
    // Let's trace it.
}
