package util

import "strings"

// IsDockerMountError checks if an error is related to Docker mounting issues, specifically overlayfs/invalid argument errors
// common in Docker-in-Docker environments.
func IsDockerMountError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "failed to mount") && (strings.Contains(msg, "overlay") || strings.Contains(msg, "invalid argument"))
}
