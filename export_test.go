package toast

import (
	"time"

	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

var HangingWrap = hangingWrap
var AlertCoords = alertCoords
var FadeIntensity = fadeIntensity

// TickMsg is the unexported tickMsg value, exposed so tests can fire a tick on
// an empty queue to exercise the defensive guard in Update.
var TickMsg = tickMsg{}

// LerpStep creates an alert with the given start/death times and returns its lerpStep value.
func LerpStep(startTime, deathTime time.Time) float64 {
	a := alert{startTime: startTime, deathTime: deathTime}
	return a.lerpStep()
}

// RenderAlert creates a minimal alert with the given parameters and renders it.
// startTime/deathTime place the alert solidly in its hold phase (full intensity).
func RenderAlert(prefix, message string, minWidth float64, maxWidth int) string {
	c, _ := colorful.Hex("#00FF00")
	a := alert{
		prefix:    prefix,
		message:   message,
		minWidth:  minWidth,
		startTime: time.Now().Add(-500 * time.Millisecond),
		deathTime: time.Now().Add(10 * time.Second),
		style:     lipgloss.NewStyle(),
		foreColor: c,
	}
	return a.render(maxWidth)
}
