package app

// isValidID checks if the ID contains only allowed characters.
// This matches the validation in util.SanitizeID but returns boolean.
func isValidID(id string) bool {
	if id == "" {
		return false
	}
	// Allow alphanumeric, underscore, hyphen.
	// Matches `[a-zA-Z0-9_-]`
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
