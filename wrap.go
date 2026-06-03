package toast

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// hangingWrap formats prefix + msg with a hanging indent so subsequent
// lines align under the first character of msg, wrapping at textWidth.
// Uses lipgloss.Width for prefix measurement to handle Unicode/NerdFont glyphs.
func hangingWrap(prefix, msg string, textWidth int) string {
	prefixWidth := lipgloss.Width(prefix)
	contentWidth := textWidth - prefixWidth - 1 // -1 for the space between prefix and msg

	if contentWidth <= 0 {
		return prefix + " " + msg
	}

	words := strings.Fields(msg)
	if len(words) == 0 {
		return prefix + " " + msg
	}

	indent := strings.Repeat(" ", prefixWidth+1)

	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
		} else {
			candidate := current.String() + " " + word
			if lipgloss.Width(candidate) <= contentWidth {
				current.WriteByte(' ')
				current.WriteString(word)
			} else {
				lines = append(lines, current.String())
				current.Reset()
				current.WriteString(word)
			}
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteByte(' ')
	sb.WriteString(lines[0])
	for _, line := range lines[1:] {
		sb.WriteByte('\n')
		sb.WriteString(indent)
		sb.WriteString(line)
	}
	return sb.String()
}
