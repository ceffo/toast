package toast_test

import (
	"testing"

	"github.com/ceffo/toast"
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
			name:   "empty message",
			prefix: "(i)", msg: "", width: 40,
			want: "(i) ",
		},
		{
			name:   "fits on one line",
			prefix: "(i)", msg: "hello", width: 20,
			want: "(i) hello",
		},
		{
			name:   "wraps at width boundary",
			prefix: "(i)", msg: "hello world", width: 10,
			// prefixWidth=3, contentWidth=10-3-1=6; "hello" fits (5), "world" does not → wrap
			// indent = prefixWidth+1 = 4 spaces
			want: "(i) hello\n    world",
		},
		{
			name:   "long message multi-wrap",
			prefix: "(!)", msg: "one two three four", width: 12,
			// prefixWidth=3, contentWidth=12-3-1=8
			// "one" (3) fits, "one two" (7) fits, "one two three" (13) → wrap at "two three"? no.
			// "one" fits, "one two" fits (7<=8), "one two three" (13>8) → lines=["one two"], current="three"
			// "three" fits (5<=8), "three four" (10>8) → lines=["one two","three"], current="four"
			// Result: "(!)\none two\n    three\n    four" — wait, leading prefix
			// "(!) one two\n    three\n    four"
			want: "(!)" + " " + "one two\n    three\n    four",
		},
		{
			name:   "content width too small falls back",
			prefix: "ⓘ", msg: "hi", width: 2,
			// prefixWidth=1 (single cell), contentWidth=2-1-1=0 → fallback
			want: "ⓘ hi",
		},
		{
			name:   "single long word that exceeds width",
			prefix: "(i)", msg: "superlongword", width: 10,
			// prefixWidth=3, contentWidth=6; word "superlongword" (13>6) but it's the only word
			// loop: current="" → current="superlongword"; end: lines=["superlongword"]
			want: "(i) superlongword",
		},
		{
			name:   "exactly fits one line",
			prefix: "(i)", msg: "hello!", width: 10,
			// contentWidth=6, "hello!" is 6 chars = exactly fits
			want: "(i) hello!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toast.HangingWrap(tc.prefix, tc.msg, tc.width)
			if got != tc.want {
				t.Errorf("HangingWrap(%q, %q, %d)\n got:  %q\n want: %q",
					tc.prefix, tc.msg, tc.width, got, tc.want)
			}
		})
	}
}
