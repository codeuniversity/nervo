package nervo

import (
	"strings"
)

// ParseAnnounceMessage parses announce messages from the controller.
// It considers messages in the form "announce <some_name>" (verb is case insensitive) ok
func ParseAnnounceMessage(line string) (name string, ok bool) {
	return parseVerb("announce", line)
}

// ParseFeedbackMessage parses feedback messages from the controller.
// It considers messages in the form "feedback <some_name>" (verb is case insensitive) ok
func ParseFeedbackMessage(line string) (message string, ok bool) {
	return parseVerb("feedback", line)
}

// ParseSensorDataMessage parses data messages from the controller
// It considers messages in the form "sensor_data <some_data>"
func ParseSensorDataMessage(line string) (message string, ok bool) {
	return parseVerb("sensor_data", line)
}

func parseVerb(verb, line string) (rest string, ok bool) {
	splitLine := strings.SplitN(line, " ", 2)
	if len(splitLine) != 2 {
		return "", false
	}

	v := strings.ToLower(splitLine[0])
	rest = removeNewLineChars(splitLine[1])

	if v != verb {
		return "", false
	}

	return rest, true
}

// ParseGaitAction parses the gait action message string into a usable leg name and message
func ParseGaitAction(line string) (legName, message string, ok bool) {
	splitLine := strings.SplitN(line, " ", 2)
	if len(splitLine) != 2 {
		return "", "", false
	}

	name := strings.ToLower(splitLine[0])
	message = removeNewLineChars(splitLine[1])

	return name, message, true
}

func removeNewLineChars(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(s, "\r\n", ""),
		"\n",
		"",
	)
}
