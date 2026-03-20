import re

with open("server/pkg/tool/types.go", "r") as f:
    content = f.read()

# Fix the stdio multiwriter type
stdout_stderr_buf_pattern = re.compile(r'_, _ = io\.Copy\(io\.MultiWriter\(stdoutBuf, &combinedBuf\), io\.LimitReader\(stdout, limit\)\)')
replacement_std = '_, _ = io.Copy(io.MultiWriter(stdoutBuf, &combinedBuf), io.LimitReader(stdout, limit))'
content = stdout_stderr_buf_pattern.sub(replacement_std, content)

stdout_stderr_buf_pattern2 = re.compile(r'_, _ = io\.Copy\(io\.MultiWriter\(stderrBuf, &combinedBuf\), io\.LimitReader\(stderr, limit\)\)')
replacement_std2 = '_, _ = io.Copy(io.MultiWriter(stderrBuf, &combinedBuf), io.LimitReader(stderr, limit))'
content = stdout_stderr_buf_pattern2.sub(replacement_std2, content)

with open("server/pkg/tool/types.go", "w") as f:
    f.write(content)
