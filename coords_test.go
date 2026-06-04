package toast

import "testing"

func Test_alertCoords(t *testing.T) {
	const cW, cH = 80, 24
	const aW, aH = 20, 4

	tests := []struct {
		pos          Position
		wantX, wantY int
	}{
		{TopLeft, 0, 0},
		{TopCenter, (cW - aW) / 2, 0},          // 30, 0
		{TopRight, cW - aW, 0},                 // 60, 0
		{BottomLeft, 0, cH - aH},               // 0, 20
		{BottomCenter, (cW - aW) / 2, cH - aH}, // 30, 20
		{BottomRight, cW - aW, cH - aH},        // 60, 20
	}

	for _, tc := range tests {
		t.Run(string(tc.pos), func(t *testing.T) {
			x, y := alertCoords(tc.pos, cW, cH, aW, aH)
			if x != tc.wantX || y != tc.wantY {
				t.Errorf("alertCoords(%s, %d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tc.pos, cW, cH, aW, aH, x, y, tc.wantX, tc.wantY)
			}
		})
	}
}

func Test_alertCoords_clampNegative(t *testing.T) {
	// Alert wider than content: x and y must not go negative.
	x, y := alertCoords(TopRight, 10, 5, 20, 10)
	if x < 0 {
		t.Errorf("x = %d, want >= 0", x)
	}
	if y < 0 {
		t.Errorf("y = %d, want >= 0", y)
	}
}
