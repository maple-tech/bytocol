package bytocol

import "strings"

// pads the ending of a string so that it meets the length provided.
// If the string is already at that length or larger it is returned unmodified.
func padStringRight(str string, length int, char rune) string {
	if len(str) >= length {
		return str
	}

	var bld strings.Builder
	bld.WriteString(str)

	rem := len(str) - length
	for i := 0; i < rem; i++ {
		bld.WriteRune(char)
	}

	return bld.String()
}
