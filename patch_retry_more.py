import re

with open('server/pkg/util/net.go', 'r') as f:
    content = f.read()

# Increase retry logic further or use a slightly longer delay to avoid OS holding it too long?
# The tests run 20 times in parallel.
# Let's change the wait to 100ms and maxRetries to 100
# Wait! `strings.Contains(errStr, "address already in use")` might not match.
# In go, bind errors look like `listen tcp 127.0.0.1:41963: bind: address already in use`
# The problem is `maxRetries := 250` and `5 * time.Millisecond` gives 1.25s.
# The context timeout is 5s.
# Let's make it `50 * time.Millisecond` and `100` retries (5s).
content = re.sub(r'backoff := 5 \* time\.Millisecond', 'backoff := 50 * time.Millisecond', content)
content = re.sub(r'maxRetries := 250', 'maxRetries := 100', content)

with open('server/pkg/util/net.go', 'w') as f:
    f.write(content)
