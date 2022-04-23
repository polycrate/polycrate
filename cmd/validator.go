package cmd

// checks if the given char is part of [.-a-zA-Z0-9]
func isValidChar(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '.' || c == '-'
}

// checks if all chars of the given string are part of [.-a-zA-Z0-9]
// could use a regex here, but direct matching should be faster, as it is a simple
// character range check
func isValidString(s string) bool {
	for _, c := range s {
		if !isValidChar(c) {
			return false
		}
	}
	return true
}
