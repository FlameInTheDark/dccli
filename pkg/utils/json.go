package utils

import (
	"encoding/json"
	"strings"
)

// ReconstructJSON attempts to fix a JSON string that might have been split by the shell
// by appending the remaining arguments.
// Returns the reconstructed JSON string and true if successful, or empty string and false if not.
func ReconstructJSON(initial string, args []string, target interface{}) (string, bool) {
	if len(args) == 0 {
		return "", false
	}

	// Try joining with spaces first, as that's the most common split cause
	joined := initial + " " + strings.Join(args, " ")
	
	// Try parsing the joined string
	if err := json.Unmarshal([]byte(joined), target); err == nil {
		return joined, true
	}
	
	// If the initial string was unquoted in shell, quotes might have been stripped.
	// But reconstructing that is much harder and risky (guessing where quotes go).
	// We primarily focus on the "space splitting" issue here.

	return "", false
}
