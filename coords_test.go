package toast_test

import (
	"testing"

	"github.com/ceffo/toast"
)

func Test_alertCoords(t *testing.T) {
	const cW, cH = 80, 24
	const aW, aH = 20, 4

	tests := []struct {
		pos          toast.Position
		wantX, wantY int
	}{
		{toast.TopLeft, 0, 0},
		{toast.TopCenter, (cW - aW) / 2, 0},          // 30, 0
		{toast.TopRight, cW - aW, 0},                 // 60, 0
		{toast.BottomLeft, 0, cH - aH},               // 0, 20
		{toast.BottomCenter, (cW - aW) / 2, cH - aH}, // 30, 20
		{toast.BottomRight, cW - aW, cH - aH},        // 60, 20
	}

	for _, tc := range tests {
		t.Run(string(tc.pos), func(t *testing.T) {
			x, y := toast.AlertCoords(tc.pos, cW, cH, aW, aH)
			if x != tc.wantX || y != tc.wantY {
				t.Errorf("AlertCoords(%s, %d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tc.pos, cW, cH, aW, aH, x, y, tc.wantX, tc.wantY)
			}
		})
	}
}

func Test_alertCoords_clampNegative(t *testing.T) {
	// Alert wider/taller than content: x and y must not go negative.
	x, y := toast.AlertCoords(toast.TopRight, 10, 5, 20, 10)
	if x < 0 {
		t.Errorf("x = %d, want >= 0", x)
	}
	if y < 0 {
		t.Errorf("y = %d, want >= 0", y)
	}
}
