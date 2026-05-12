package commands

import "strings"

// ParseCommand parses a slash command and returns command name and args
func ParseCommand(input string) (cmd string, args string, isCommand bool) {
	if len(input) == 0 || input[0] != '/' {
		return "", "", false
	}

	input = input[1:] // Remove leading /

	// Handle commands with no arguments
	if input == "" {
		return "", "", false
	}

	parts := strings.SplitN(input, " ", 2)

	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return parts[0], "", true
}

// FindCommandMatch checks if input matches a command (by name or alias)
func FindCommandMatch(name string, aliases []string, search string) bool {
	if name == search {
		return true
	}
	for _, alias := range aliases {
		if alias == search {
			return true
		}
	}
	return false
}
