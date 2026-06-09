package toast

import (
	"time"

	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type alert struct {
	message   string
	startTime time.Time
	deathTime time.Time
	prefix    string
	foreColor colorful.Color
	style     lipgloss.Style
	minWidth  float64
	position  Position
}

// lerpStep returns a value in [0,1] representing how far through the alert's
// lifetime we are, derived from startTime and deathTime at call time.
func (a alert) lerpStep() float64 {
	duration := a.deathTime.Sub(a.startTime)
	if duration <= 0 {
		return 1
	}
	step := float64(time.Since(a.startTime)) / float64(duration)
	if step < 0 {
		return 0
	}
	if step > 1 {
		return 1
	}
	return step
}

// fadeIntensity maps a lifetime progress value (0→1) to a color-blend factor
// with three phases:
//
//	0.00–0.15  fade in  (0→1)
//	0.15–0.75  hold     (1)
//	0.75–1.00  fade out (1→0)
func fadeIntensity(step float64) float64 {
	const fadeIn = 0.15
	const fadeOut = 0.75
	switch {
	case step < fadeIn:
		return step / fadeIn
	case step > fadeOut:
		return 1.0 - (step-fadeOut)/(1.0-fadeOut)
	default:
		return 1.0
	}
}

func (a alert) render(maxWidth int) string {
	// Lab-space lerp: foreColor → black over the last 25% of the lifetime.
	blended := a.foreColor.BlendLab(colorful.Color{}, 1.0-fadeIntensity(a.lerpStep()))
	col := lipgloss.Color(blended.Hex())

	// Inner content area: border (1 each side) + padding (1 each side) = 4 total overhead
	maxW := maxWidth
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
