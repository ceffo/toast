package toast

import (
	"time"

	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type alert struct {
	message     string
	deathTime   time.Time
	prefix      string
	foreColor   colorful.Color
	style       lipgloss.Style
	width       float64
	minWidth    float64
	curLerpStep float64
	position    Position
}

func (a alert) render() string {
	// Lab-space lerp: black → foreColor
	blended := colorful.Color{}.BlendLab(a.foreColor, a.curLerpStep)
	col := lipgloss.Color(blended.Hex())

	// Inner content area: border (1 each side) + padding (1 each side) = 4 total overhead
	maxW := int(a.width)
	innerWidth := maxW - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	content := hangingWrap(a.prefix, a.message, innerWidth)

	actualWidth := maxW
	if a.minWidth > 0 {
		natural := lipgloss.Width(content) + 4
		minW := int(a.minWidth)
		if natural < minW {
			natural = minW
		}
		if natural > maxW {
			natural = maxW
		}
		actualWidth = natural
	}

	s := a.style.
		Border(lipgloss.RoundedBorder()).
		Foreground(col).
		BorderForeground(col).
		Width(actualWidth).
		Padding(0, 1)

	return s.Render(content)
}
