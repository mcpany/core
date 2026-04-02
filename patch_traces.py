import re

with open('server/pkg/app/api_traces.go', 'r') as f:
    content = f.read()

# Fix the multiline string literal
old_str = '''"diff":   "--- a/main.py
+++ b/main.py
@@ -1,5 +1,5 @@
-def slow_func():
-    pass
+def fast_func():
+    return True
",'''

new_str = '"diff": "--- a/main.py\\n+++ b/main.py\\n@@ -1,5 +1,5 @@\\n-def slow_func():\\n-    pass\\n+def fast_func():\\n+    return True\\n",'

content = content.replace(old_str, new_str)

with open('server/pkg/app/api_traces.go', 'w') as f:
    f.write(content)
