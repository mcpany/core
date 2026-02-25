# Copyright 2025 Author(s) of MCP Any
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM mirror.gcr.io/library/python:3.11-slim

# Use /app as working directory
WORKDIR /app

# Install dependencies directly to minimize layer issues
RUN pip install "fastmcp>=2.0.0"

# Inline main.py content to avoid COPY context mount issues on some CI runners
# Content is base64 encoded to preserve formatting and avoid escaping hell
RUN echo "IyEvdXNyL2Jpbi9lbnYgcHl0aG9uMwojIENvcHlyaWdodCAyMDI1IEF1dGhvcihzKSBvZiBNQ1AgQW55CiMgU1BEWC1MaWNlbnNlLUlkZW50aWZpZXI6IEFwYWNoZS0yLjAKCgpmcm9tIGRhdGV0aW1lIGltcG9ydCBkYXRldGltZQppbXBvcnQgc3lzCgpmcm9tIGZhc3RtY3AgaW1wb3J0IHN0YXJ0X21jcF9zZXJ2ZXIKCmRlZiBnZXRfbG9jYWxfdGltZV9qc29uKGFyZ3MpOgogICAgIiIiQSBzaW1wbGUgdG9vbCB0aGF0IHJldHVybnMgdGhlIGN1cnJlbnQgdGltZSBhcyBKU09OLiIiIgogICAgbm93ID0gZGF0ZXRpbWUubm93KCkuYXN0aW1lem9uZSgpCiAgICB0aW1lX3N0ciA9IG5vdy5zdHJmdGltZSgiJVktJW0tJWQgJUg6JU06JVMiKQogICAgdGltZXpvbmVfc3RyID0gbm93LnR6bmFtZSgpCiAgICByZXR1cm4gewogICAgICAgICJjdXJyZW50X3RpbWUiOiB0aW1lX3N0ciwKICAgICAgICAidGltZXpvbmUiOiB0aW1lem9uZV9zdHIKICAgIH0KCmRlZiBtYWluKCk6CiAgICAiIiIKICAgIE1haW4gZW50cnkgcG9pbnQgZm9yIHRoZSB0aW1lIHNlcnZlci4KICAgICIiIgogICAgc3RhcnRfbWNwX3NlcnZlcigKICAgICAgICBzZXJ2aWNlX25hbWU9InRpbWUtc2VydmVyIiwKICAgICAgICBzZXJ2aWNlX2Rlc2NyaXB0aW9uPSJBIHNpbXBsZSBzZXJ2aWNlIHRoYXQgcHJvdmlkZXMgdGhlIGN1cnJlbnQgdGltZS4iLAogICAgICAgIHRvb2xzPXsiZ2V0X2xvY2FsX3RpbWUiOiBnZXRfbG9jYWxfdGltZV9qc29ufSwKICAgICAgICBpbnB1dF9zdHJlYW09c3lzLnN0ZGluLAogICAgICAgIG91dHB1dF9zdHJlYW09c3lzLnN0ZG91dAogICAgKQoKaWYgX19uYW1lX18gPT0gIl9fbWFpbl9fIjoKICAgIG1haW4oKQo=" | base64 -d > main.py

CMD ["python", "-u", "main.py", "--mcp-stdio"]
