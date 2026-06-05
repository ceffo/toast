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

// fadeIntensity maps curLerpStep (0→1 over the alert's lifetime) to a
// color-blend factor with three phases:
//
//	0.00–0.15  fade in  (0→1)
//	0.15–0.75  hold     (1)
//	0.75–1.00  fade out (1→0)
func (a alert) fadeIntensity() float64 {
	const fadeIn = 0.15
	const fadeOut = 0.75
	switch {
	case a.curLerpStep < fadeIn:
		return a.curLerpStep / fadeIn
	case a.curLerpStep > fadeOut:
		return 1.0 - (a.curLerpStep-fadeOut)/(1.0-fadeOut)
	default:
		return 1.0
	}
}

func (a alert) render() string {
	// Lab-space lerp: foreColor → black over the last 25% of the lifetime.
	blended := a.foreColor.BlendLab(colorful.Color{}, 1.0-a.fadeIntensity())
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
