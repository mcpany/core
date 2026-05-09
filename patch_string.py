import re

with open('server/pkg/util/string.go', 'r') as f:
    content = f.read()

replacement = """	for i := 1; i <= n; i++ {
		v1[0] = i
		minRow := v1[0]
		for j := 1; j <= m; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			v1[j] = min(
				v0[j]+1,      // deletion
				v1[j-1]+1,    // insertion
				v0[j-1]+cost, // substitution
			)
			if v1[j] < minRow {
				minRow = v1[j]
			}
		}

		// If the minimum value in this row exceeds the limit, we can stop early.
		if minRow > limit {
			return limit + 1
		}

		// Swap slices for next iteration
		for k := 0; k <= m; k++ {
			v0[k] = v1[k]
		}
	}"""

pattern = r"""	for i := 1; i <= n; i\+\+ \{
		v1\[0\] = i
		minRow := v1\[0\]
		for j := 1; j <= m; j\+\+ \{
			cost := 0
			if s1\[i-1\] != s2\[j-1\] \{
				cost = 1
			\}
			v1\[j\] = min\(
				v0\[j\]\+1,      // deletion
				v1\[j-1\]\+1,    // insertion
				v0\[j-1\]\+cost, // substitution
			\)
			if v1\[j\] < minRow \{
				minRow = v1\[j\]
			\}
		\}

		// If the minimum value in this row exceeds the limit, we can stop early\.
		if minRow > limit \{
			return limit \+ 1
		\}

		// Swap v0 and v1
		v0, v1 = v1, v0
	\}"""

content = re.sub(pattern, replacement, content)

with open('server/pkg/util/string.go', 'w') as f:
    f.write(content)
