import re

with open("server/pkg/dmr/hub.go", "r") as f:
    content = f.read()

def replace_func(m):
    func_sig = m.group(0)
    func_name = m.group(1)

    if func_name == "RegisterNode":
        doc = f"""\t// Summary: Registers a node in the DMR Hub.
\t//
\t// Parameters:
\t//   - id (string): The unique identifier of the node.
\t//   - isAttested (bool): Whether the node has provided valid hardware attestation.
\t//
\t// Returns:
\t//   - error: An error if registration fails.
\t//
\t// Errors:
\t//   - Returns error if id is invalid.
\t//
\t// Side Effects:
\t//   - Modifies the internal nodes map."""
    elif func_name == "Heartbeat":
        doc = f"""\t// Summary: Processes a heartbeat signal from a mesh node.
\t//
\t// Parameters:
\t//   - id (string): The unique identifier of the node.
\t//
\t// Returns:
\t//   - error: An error if the node is not found.
\t//
\t// Errors:
\t//   - Returns error if the node is not registered.
\t//
\t// Side Effects:
\t//   - Updates the LastHeartbeat time for the node."""
    elif func_name == "CheckHealth":
        doc = f"""\t// Summary: Evaluates node health and triggers migration for failed nodes.
\t//
\t// Parameters:
\t//   - ctx (context.Context): The context for the health check.
\t//
\t// Returns:
\t//   - []string: A list of node IDs that have failed and require migration.
\t//
\t// Errors:
\t//   - None.
\t//
\t// Side Effects:
\t//   - Can send failed node IDs to the migration channel."""
    elif func_name == "MigrationChannel":
        doc = f"""\t// Summary: Provides access to the stream of failed node IDs that require migration.
\t//
\t// Parameters:
\t//   - None.
\t//
\t// Returns:
\t//   - <-chan string: A channel emitting failed node IDs.
\t//
\t// Errors:
\t//   - None.
\t//
\t// Side Effects:
\t//   - None."""
    else:
        return m.group(0) # Unchanged if not matched

    return doc + "\n" + func_sig.split('\n')[-1]

content = re.sub(r'\t// Summary: [^\n]*\n\t(\w+)\(.*?\).*?', lambda m: replace_func(m), content)

with open("server/pkg/dmr/hub.go", "w") as f:
    f.write(content)
