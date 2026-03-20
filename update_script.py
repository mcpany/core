import re

with open("server/pkg/tool/types.go", "r") as f:
    content = f.read()

# Add sync.Pool for bytes.Buffer
sync_pool_declaration = """
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// getBuffer returns a bytes.Buffer from the pool.
func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer returns a bytes.Buffer to the pool after resetting it.
func putBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
"""

# Find the right place to insert it. After `var fastJSONNumber` definition.
fastJSON_pattern = re.compile(r'var fastJSONNumber = jsoniter\.Config\{\n\tEscapeHTML:             true,\n\tSortMapKeys:            true,\n\tValidateJsonRawMessage: true,\n\tUseNumber:              true,\n\}\.Froze\(\)\n', re.MULTILINE)

match = fastJSON_pattern.search(content)
if match:
    content = content[:match.end()] + "\n" + sync_pool_declaration + "\n" + content[match.end():]

# Now let's replace `var headerBuf bytes.Buffer`
header_buf_pattern_1 = re.compile(r'\t// Log headers\n\tvar headerBuf bytes\.Buffer\n')
content = header_buf_pattern_1.sub('\t// Log headers\n\t// ⚡ BOLT: Replaced heap-allocated bytes.Buffer with sync.Pool for O(1) memory reuse under high throughput HTTP traffic.\n\t// Randomized Selection from Top 5 High-Impact Targets (Memory)\n\theaderBuf := getBuffer()\n\tdefer putBuffer(headerBuf)\n', content)

header_buf_pattern_2 = re.compile(r'\t\t// Log headers\n\t\tvar headerBuf bytes\.Buffer\n')
content = header_buf_pattern_2.sub('\t\t// Log headers\n\t\t// ⚡ BOLT: Replaced heap-allocated bytes.Buffer with sync.Pool for O(1) memory reuse under high throughput HTTP traffic.\n\t\t// Randomized Selection from Top 5 High-Impact Targets (Memory)\n\t\theaderBuf := getBuffer()\n\t\tdefer putBuffer(headerBuf)\n', content)

# Replace `var stderrBuf bytes.Buffer` in LocalCommandTool & CommandTool
# First, the JSON stdio case
stderr_buf_json_pattern = re.compile(r'\t\tvar stderrBuf bytes\.Buffer\n\t\tstderrDone := make\(chan struct\{\}\)\n\t\tgo func\(\) \{\n\t\t\tdefer close\(stderrDone\)\n\t\t\tdefer func\(\) \{ _ = stderr\.Close\(\) \}\(\)\n\t\t\t_, _ = io\.Copy\(&stderrBuf, io\.LimitReader\(stderr, limit\)\)\n\t\t\}\(\)\n')

replacement_json = """\t\t// ⚡ BOLT: Replaced heap-allocated bytes.Buffer with sync.Pool for O(1) memory reuse during command execution streams.
\t\t// Randomized Selection from Top 5 High-Impact Targets (Memory)
\t\tstderrBuf := getBuffer()
\t\tdefer putBuffer(stderrBuf)
\t\tstderrDone := make(chan struct{})
\t\tgo func() {
\t\t\tdefer close(stderrDone)
\t\t\tdefer func() { _ = stderr.Close() }()
\t\t\t_, _ = io.Copy(stderrBuf, io.LimitReader(stderr, limit))
\t\t}()
"""
content = stderr_buf_json_pattern.sub(replacement_json, content)

# Second, the standard pipe case
stdout_stderr_buf_pattern = re.compile(r'\tvar stdoutBuf, stderrBuf bytes\.Buffer\n\tvar combinedBuf threadSafeBuffer\n\tvar wg sync\.WaitGroup\n\twg\.Add\(2\)\n\n\tgo func\(\) \{\n\t\tdefer wg\.Done\(\)\n\t\tdefer func\(\) \{ _ = stdout\.Close\(\) \}\(\)\n\t\t_, _ = io\.Copy\(io\.MultiWriter\(&stdoutBuf, &combinedBuf\), io\.LimitReader\(stdout, limit\)\)\n\t\}\(\)\n\tgo func\(\) \{\n\t\tdefer wg\.Done\(\)\n\t\tdefer func\(\) \{ _ = stderr\.Close\(\) \}\(\)\n\t\t_, _ = io\.Copy\(io\.MultiWriter\(&stderrBuf, &combinedBuf\), io\.LimitReader\(stderr, limit\)\)\n\t\}\(\)\n')

replacement_std = """\t// ⚡ BOLT: Replaced heap-allocated bytes.Buffer with sync.Pool for O(1) memory reuse during command execution streams.
\t// Randomized Selection from Top 5 High-Impact Targets (Memory)
\tstdoutBuf := getBuffer()
\tdefer putBuffer(stdoutBuf)
\tstderrBuf := getBuffer()
\tdefer putBuffer(stderrBuf)

\tvar combinedBuf threadSafeBuffer
\tvar wg sync.WaitGroup
\twg.Add(2)

\tgo func() {
\t\tdefer wg.Done()
\t\tdefer func() { _ = stdout.Close() }()
\t\t_, _ = io.Copy(io.MultiWriter(stdoutBuf, &combinedBuf), io.LimitReader(stdout, limit))
\t}()
\tgo func() {
\t\tdefer wg.Done()
\t\tdefer func() { _ = stderr.Close() }()
\t\t_, _ = io.Copy(io.MultiWriter(stderrBuf, &combinedBuf), io.LimitReader(stderr, limit))
\t}()
"""
content = stdout_stderr_buf_pattern.sub(replacement_std, content)


# Now fix the reference to stderrBuf.String() in JSON decode error
content = content.replace("stderrBuf.String()", "stderrBuf.String()")


# And replacing the return lines referencing stdoutBuf and stderrBuf strings.
content = content.replace("stdoutBuf.String()", "stdoutBuf.String()")
content = content.replace("stderrBuf.String()", "stderrBuf.String()")


# For prettyPrint xml
xml_buf_pattern = re.compile(r'\t\tdecoder := xml\.NewDecoder\(bytes\.NewReader\(input\)\)\n\t\tvar buf bytes\.Buffer\n\t\tencoder := xml\.NewEncoder\(&buf\)\n')
replacement_xml = """\t\tdecoder := xml.NewDecoder(bytes.NewReader(input))
\t\t// ⚡ BOLT: Replaced heap-allocated bytes.Buffer with sync.Pool to minimize GC pressure during intensive logging operations.
\t\t// Randomized Selection from Top 5 High-Impact Targets (Memory)
\t\tbuf := getBuffer()
\t\tdefer putBuffer(buf)
\t\tencoder := xml.NewEncoder(buf)
"""
content = xml_buf_pattern.sub(replacement_xml, content)

# For prettyJSON
json_buf_pattern = re.compile(r'\t\tvar prettyJSON bytes\.Buffer\n\t\t// Use stdjson for Indent\n\t\tif err := stdjson\.Indent\(&prettyJSON, input, "", "  "\); err == nil \{\n\t\t\treturn prettyJSON\.String\(\)\n\t\t\}\n')
replacement_json_print = """\t\t// ⚡ BOLT: Replaced heap-allocated bytes.Buffer with sync.Pool to minimize GC pressure during intensive logging operations.
\t\t// Randomized Selection from Top 5 High-Impact Targets (Memory)
\t\tprettyJSON := getBuffer()
\t\tdefer putBuffer(prettyJSON)
\t\t// Use stdjson for Indent
\t\tif err := stdjson.Indent(prettyJSON, input, "", "  "); err == nil {
\t\t\treturn prettyJSON.String()
\t\t}
"""
content = json_buf_pattern.sub(replacement_json_print, content)

with open("server/pkg/tool/types.go", "w") as f:
    f.write(content)
