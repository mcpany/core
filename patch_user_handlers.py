import re

with open('server/pkg/app/user_handlers.go', 'r') as f:
    content = f.read()

# Replace `user = configv1.User_builder{` with `user = (&configv1.User_builder{`
content = content.replace('user = configv1.User_builder{', 'user = (&configv1.User_builder{')

# Also replace `.Build()` with `}).Build()`
# This specifically occurs a few lines after.
import re
content = re.sub(r'user = \(&configv1\.User_builder\{(.*?)\}\.Build\(\)', r'user = (&configv1.User_builder{\1}).Build()', content, flags=re.DOTALL)

with open('server/pkg/app/user_handlers.go', 'w') as f:
    f.write(content)
