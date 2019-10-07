package nervo

import (
	"strings"
)

// ParseAnnounceMessage parses messages announce messages from the controller.
// It considers messages in the form "announce <some_name>" (verb is case insensitive) ok
func ParseAnnounceMessage(line string) (name string, ok bool) {
	splitLine := strings.SplitN(line, " ", 2)
	if len(splitLine) != 2 {
		return "", false
	}

	verb := strings.ToLower(splitLine[0])
	name = removeNewLineChars(splitLine[1])

	if verb != "announce" {
		return "", false
	}

	return name, true
}

func removeNewLineChars(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(s, "\r\n", ""),
		"\n",
		"",
	)
}
