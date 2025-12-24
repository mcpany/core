# Mini Design Doc: Filesystem Search & Delete

## Objective
Enhance the Filesystem Provider with `search_files` and `delete_file` capabilities to improve utility for AI agents.

## New Tools

### 1. `search_files`
**Description**: Recursively search for a text pattern in files within a directory.

**Input**:
- `path` (string): The root directory to start searching from.
- `pattern` (string): The regular expression pattern to search for.
- `recursive` (boolean, optional): Whether to search recursively. Default: `true`. (Wait, strict recursiveness might be expensive. Let's make it optional but default false for safety? Or default true but limited depth? Let's say default true but we limit total matches).
- `exclude_patterns` (array of strings, optional): Patterns to exclude (e.g., `node_modules`, `.git`).

**Output**:
- `matches`: Array of objects:
  - `file`: File path relative to root.
  - `line_number`: Line number.
  - `line_content`: Content of the matching line (trimmed).

**Constraints**:
- Max matches limit (e.g., 100).
- Max file size to read (e.g., 10MB).
- Binary file detection (skip binary files).

**Edge Cases**:
- Invalid regex.
- Permission denied on subdirectories.
- Symlink loops (afero `Walk` should handle or we need to be careful).

### 2. `delete_file`
**Description**: Delete a file or empty directory.

**Input**:
- `path` (string): The path to delete.

**Output**:
- `success` (boolean).

**Constraints**:
- Respect `ReadOnly` flag in `FilesystemUpstreamService`.
- Recursive delete? Maybe safer to only allow single file or empty dir for now. Or add `recursive` flag for directories. Let's stick to `fs.Remove` which is non-recursive for directories usually, or `RemoveAll` if we want recursive. Let's do `Remove` for safety first, or maybe `delete_file` implies file.
- `afero.Fs.Remove` removes a file or an empty directory. `RemoveAll` removes path and any children.
- Let's expose `delete_file` (using `Remove`) and maybe `delete_tree` later.

## Implementation Details

- **File**: `server/pkg/upstream/filesystem/upstream.go`
- **Dependencies**: `regexp`, `bufio`, `unicode/utf8` (for binary check).

### Binary Check
Read first 512 bytes. If `http.DetectContentType` says "application/octet-stream" or if we find null bytes, skip.

### Search Logic
1. Resolve path.
2. `afero.Walk`.
3. Check exclusions.
4. If file, check size.
5. Read file. Check binary.
6. Scan lines. Match regex.
7. Collect matches. Stop if limit reached.

