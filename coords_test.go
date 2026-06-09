package toast_test

import (
	"testing"

	"github.com/ceffo/toast"
	"github.com/stretchr/testify/assert"
)

func TestAlertCoords(t *testing.T) {
	const cW, cH = 80, 24
	const aW, aH = 20, 4

	tests := []struct {
		pos          toast.Position
		wantX, wantY int
	}{
		{pos: toast.TopLeft, wantX: 0, wantY: 0},
		{pos: toast.TopCenter, wantX: (cW - aW) / 2, wantY: 0},
		{pos: toast.TopRight, wantX: cW - aW, wantY: 0},
		{pos: toast.BottomLeft, wantX: 0, wantY: cH - aH},
		{pos: toast.BottomCenter, wantX: (cW - aW) / 2, wantY: cH - aH},
		{pos: toast.BottomRight, wantX: cW - aW, wantY: cH - aH},
	}

	for _, tc := range tests {
		t.Run(string(tc.pos), func(t *testing.T) {
			x, y := toast.AlertCoords(tc.pos, cW, cH, aW, aH)
			assert.Equal(t, tc.wantX, x)
			assert.Equal(t, tc.wantY, y)
		})
	}
}

func TestAlertCoords_clampNegative(t *testing.T) {
	tests := []struct {
		name               string
		pos                toast.Position
		contentW, contentH int
		alertW, alertH     int
	}{
		{
			name:     "alert_wider_than_content_top_right",
			pos:      toast.TopRight,
			contentW: 10, contentH: 5,
			alertW: 20, alertH: 10,
		},
		{
			name:     "alert_taller_than_content_bottom_left",
			pos:      toast.BottomLeft,
			contentW: 80, contentH: 2,
			alertW: 20, alertH: 10,
		},
		{
			name:     "alert_larger_in_both_dimensions",
			pos:      toast.BottomRight,
			contentW: 5, contentH: 3,
			alertW: 20, alertH: 10,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, y := toast.AlertCoords(tc.pos, tc.contentW, tc.contentH, tc.alertW, tc.alertH)
			assert.GreaterOrEqual(t, x, 0, "x must not be negative")
			assert.GreaterOrEqual(t, y, 0, "y must not be negative")
		})
	}
}
