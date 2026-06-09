package toast_test

import (
	"testing"

	"github.com/ceffo/toast"
	"github.com/stretchr/testify/assert"
)

func TestAlertDefinition_Resolve(t *testing.T) {
	def := toast.AlertDefinition{
		Prefix:    ">>",
		ForeColor: "#FF0000",
		Position:  toast.TopLeft,
	}
	fonts := []toast.FontStyle{toast.FontASCII, toast.FontUnicode, toast.FontNerdFont}
	for _, font := range fonts {
		t.Run(string(font), func(t *testing.T) {
			assert.Equal(t, def, def.Resolve(font), "AlertDefinition.Resolve must return itself regardless of font style")
		})
	}
}

func TestAlertLevel_Resolve(t *testing.T) {
	tests := []struct {
		name       string
		font       toast.FontStyle
		wantPrefix string
		wantColor  string
	}{
		{name: "ascii", font: toast.FontASCII, wantPrefix: "(i)", wantColor: "#00FF00"},
		{name: "unicode", font: toast.FontUnicode, wantPrefix: "ⓘ", wantColor: "#00FF00"},
		{name: "nerdfont", font: toast.FontNerdFont, wantPrefix: "\U000f02fc", wantColor: "#00FF00"},
		{name: "unknown_falls_back_to_ascii", font: toast.FontStyle("unknown"), wantPrefix: "(i)", wantColor: "#00FF00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			def := toast.InfoAlert.Resolve(tc.font)
			assert.Equal(t, tc.wantPrefix, def.Prefix)
			assert.Equal(t, tc.wantColor, def.ForeColor)
		})
	}
}

func TestAlertLevel_Resolve_allBuiltinLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     toast.AlertLevel
		wantASCII string
		wantColor string
	}{
		{name: "info", level: toast.InfoAlert, wantASCII: "(i)", wantColor: "#00FF00"},
		{name: "warn", level: toast.WarnAlert, wantASCII: "(!)", wantColor: "#FFFF00"},
		{name: "error", level: toast.ErrorAlert, wantASCII: "[!!]", wantColor: "#FF0000"},
		{name: "debug", level: toast.DebugAlert, wantASCII: "(?)", wantColor: "#FF00FF"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			def := tc.level.Resolve(toast.FontASCII)
			assert.Equal(t, tc.wantASCII, def.Prefix)
			assert.Equal(t, tc.wantColor, def.ForeColor)
		})
	}
}

func TestPosition_IsValid(t *testing.T) {
	tests := []struct {
		name string
		pos  toast.Position
		want bool
	}{
		{name: "top-left", pos: toast.TopLeft, want: true},
		{name: "top-center", pos: toast.TopCenter, want: true},
		{name: "top-right", pos: toast.TopRight, want: true},
		{name: "bottom-left", pos: toast.BottomLeft, want: true},
		{name: "bottom-center", pos: toast.BottomCenter, want: true},
		{name: "bottom-right", pos: toast.BottomRight, want: true},
		{name: "unspecified/zero_value", pos: toast.UnspecifiedPosition, want: false},
		{name: "arbitrary_string", pos: toast.Position("bogus"), want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.pos.IsValid())
		})
	}
}
