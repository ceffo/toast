package toast_test

import (
	"testing"

	"github.com/ceffo/toast"
	"github.com/stretchr/testify/assert"
)

func TestHangingWrap(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		msg    string
		width  int
		want   string
	}{
		{
			name:   "empty_message",
			prefix: "(i)", msg: "", width: 40,
			want: "(i) ",
		},
		{
			name:   "fits_on_one_line",
			prefix: "(i)", msg: "hello", width: 20,
			want: "(i) hello",
		},
		{
			name:   "wraps_at_width_boundary",
			prefix: "(i)", msg: "hello world", width: 10,
			// prefixWidth=3, contentWidth=10-3-1=6
			// "hello"(5) fits, "hello world"(11) does not → wrap
			// indent = 3+1 = 4 spaces
			want: "(i) hello\n    world",
		},
		{
			name:   "multi_wrap",
			prefix: "(!)", msg: "one two three four", width: 12,
			// prefixWidth=3, contentWidth=12-3-1=8
			// "one two"(7)≤8 fits; "one two three"(13)>8 → line "one two"
			// "three"(5)≤8; "three four"(10)>8 → line "three"
			// remaining: "four"
			want: "(!)" + " " + "one two\n    three\n    four",
		},
		{
			name:   "content_width_too_small_falls_back",
			prefix: "ⓘ", msg: "hi", width: 2,
			// prefixWidth=1, contentWidth=2-1-1=0 → fallback (no wrapping)
			want: "ⓘ hi",
		},
		{
			name:   "single_word_exceeds_content_width",
			prefix: "(i)", msg: "superlongword", width: 10,
			// prefixWidth=3, contentWidth=6; word(13)>6 but only word → placed as-is
			want: "(i) superlongword",
		},
		{
			name:   "exactly_fits_one_line",
			prefix: "(i)", msg: "hello!", width: 10,
			// prefixWidth=3, contentWidth=6; "hello!"(6)==6 fits
			want: "(i) hello!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toast.HangingWrap(tc.prefix, tc.msg, tc.width)
			assert.Equal(t, tc.want, got)
		})
	}
}
