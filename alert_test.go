package toast_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ceffo/toast"
	"github.com/stretchr/testify/assert"
)

func TestFadeIntensity(t *testing.T) {
	tests := []struct {
		name string
		step float64
		want float64
	}{
		// Fade-in phase: step < 0.15 → intensity = step/0.15
		{name: "zero_is_fully_transparent", step: 0.00, want: 0.0},
		{name: "mid_fade_in", step: 0.075, want: 0.5},
		// Boundary: step == 0.15 is NOT < 0.15, so falls into hold phase
		{name: "at_fade_in_boundary_holds", step: 0.15, want: 1.0},
		// Hold phase: 0.15 ≤ step ≤ 0.75 → intensity = 1
		{name: "hold_mid", step: 0.45, want: 1.0},
		// Boundary: step == 0.75 is NOT > 0.75, so still hold phase
		{name: "at_fade_out_boundary_holds", step: 0.75, want: 1.0},
		// Fade-out phase: step > 0.75 → intensity = 1 - (step-0.75)/0.25
		{name: "mid_fade_out", step: 0.875, want: 0.5},
		{name: "at_one_fully_faded", step: 1.0, want: 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toast.FadeIntensity(tc.step)
			assert.InDelta(t, tc.want, got, 1e-9)
		})
	}
}

func TestLerpStep(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		startTime time.Time
		deathTime time.Time
		want      float64
	}{
		{
			name:      "zero_duration_returns_1",
			startTime: now,
			deathTime: now,
			want:      1.0,
		},
		{
			name:      "negative_duration_returns_1",
			startTime: now.Add(time.Second),
			deathTime: now, // deathTime before startTime → duration < 0
			want:      1.0,
		},
		{
			name:      "future_start_clamps_to_0",
			startTime: now.Add(time.Hour),
			deathTime: now.Add(2 * time.Hour),
			want:      0.0,
		},
		{
			name:      "past_death_clamps_to_1",
			startTime: now.Add(-10 * time.Second),
			deathTime: now.Add(-1 * time.Second), // both in the past, alert long expired
			want:      1.0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toast.LerpStep(tc.startTime, tc.deathTime)
			assert.InDelta(t, tc.want, got, 1e-6)
		})
	}
}

func TestRenderAlert(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		message  string
		minWidth float64
		maxWidth int
	}{
		{
			name:   "no_min_width",
			prefix: "(i)", message: "hello world",
			minWidth: 0, maxWidth: 40,
		},
		{
			name:   "min_width_below_natural_size",
			prefix: "(i)", message: "this is a fairly long message here",
			minWidth: 5, maxWidth: 40,
		},
		{
			name:   "min_width_above_natural_size",
			prefix: "(i)", message: "hi",
			minWidth: 30, maxWidth: 40,
		},
		{
			name:   "min_width_exceeds_max_clamped_to_max",
			prefix: "(i)", message: "hi",
			minWidth: 60, maxWidth: 40,
		},
		{
			name:   "very_small_max_width_no_panic",
			prefix: "(i)", message: "hi",
			minWidth: 0, maxWidth: 2,
		},
		{
			name:   "unicode_prefix",
			prefix: "ⓘ", message: "info message",
			minWidth: 0, maxWidth: 40,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toast.RenderAlert(tc.prefix, tc.message, tc.minWidth, tc.maxWidth)
			assert.NotEmpty(t, got)
			// Message may be word-wrapped across lines; check a leading word instead.
			firstWord := strings.Fields(tc.message)[0]
			assert.Contains(t, got, firstWord)
		})
	}
}
