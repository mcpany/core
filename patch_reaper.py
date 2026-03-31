import re

with open("server/pkg/lifecycle/reaper.go", "r") as f:
    content = f.read()

def replace_func(m):
    func_sig = m.group(0)
    func_name = m.group(1)

    if func_name == "NewSubagentReaper":
        doc = f"""// {func_name} initializes a new Active Subagent Reaper.
//
// Summary: Creates a new SubagentReaper instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *SubagentReaper: The initialized reaper.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Initializes internal channels and maps."""
    elif func_name == "RegisterBranch":
        doc = f"""// {func_name} creates a new speculative intent branch and assigns a lease.
//
// Summary: Registers a new intent branch with a lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//   - ttl (time.Duration): The time-to-live for the lease.
//
// Returns:
//   - *Lease: The newly created lease.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the internal registry map."""
    elif func_name == "RegisterSubagent":
        doc = f"""// {func_name} attaches a subagent session to a lease.
//
// Summary: Registers a subagent with an existing intent lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//   - sessionID (string): The unique identifier for the subagent session.
//
// Returns:
//   - error: An error if the lease is not found or not active.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//   - Returns "cannot register subagent: lease is [status]" if not active.
//
// Side Effects:
//   - Modifies the lease's SubagentSessionIDs list."""
    elif func_name == "RecordHeartbeat":
        doc = f"""// {func_name} updates the Last-Seen timestamp (Expiry) for an active lease.
//
// Summary: Extends the expiration time of an active lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//   - signature (string): The heartbeat signature.
//   - extendBy (time.Duration): The duration to extend the lease by.
//
// Returns:
//   - error: An error if the lease is not found or not active.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//   - Returns "cannot record heartbeat: lease is [status]" if not active.
//
// Side Effects:
//   - Modifies the Expiry timestamp of the lease."""
    elif func_name == "PruneIntent":
        doc = f"""// {func_name} manually invalidates a lease and rolls back uncommitted writes.
//
// Summary: Marks an intent lease as pruned.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//
// Returns:
//   - error: An error if the lease is not found.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//
// Side Effects:
//   - Changes the lease status to StatusPruned."""
    elif func_name == "Start":
        doc = f"""// {func_name} begins the Reaper Daemon background worker.
//
// Summary: Starts the background process to sweep expired leases.
//
// Parameters:
//   - ctx (context.Context): The context to control the daemon lifecycle.
//   - interval (time.Duration): The interval between sweep operations.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Starts a new goroutine for the daemon worker."""
    elif func_name == "Stop":
        doc = f"""// {func_name} halts the Reaper Daemon.
//
// Summary: Stops the background process sweeping expired leases.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Closes the quit channel to signal the daemon to stop."""
    elif func_name == "GetLeaseStatus":
        doc = f"""// {func_name} returns the status of a given lease.
//
// Summary: Retrieves the current status of an intent lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//
// Returns:
//   - LeaseStatus: The current status of the lease.
//   - error: An error if the lease is not found.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//
// Side Effects:
//   - None."""
    else:
        return m.group(0) # Unchanged if not matched

    return doc + "\n" + func_sig.split('\n')[-1]

# Need to replace lines like "// NewSubagentReaper initializes..." with doc block
content = re.sub(r'// (\w+) [^\n]*\nfunc \([^)]+\) \1\(.*?\).*?\{|// (\w+) [^\n]*\nfunc \1\(.*?\).*?\{', lambda m: replace_func(m), content)

with open("server/pkg/lifecycle/reaper.go", "w") as f:
    f.write(content)
