package toast

// FontStyle selects the glyph set used for alert prefixes.
type FontStyle string

const (
	FontASCII    FontStyle = "ascii"
	FontUnicode  FontStyle = "unicode"
	FontNerdFont FontStyle = "nerdfont"
)

// AlertSpec is the common interface for anything that can be passed to
// NewAlertCmd. AlertDefinition satisfies it by returning itself; AlertLevel
// satisfies it by resolving to the right variant for the model's FontStyle.
type AlertSpec interface {
	Resolve(FontStyle) AlertDefinition
}

// AlertDefinition holds the visual configuration for one alert level.
type AlertDefinition struct {
	Prefix    string
	ForeColor string
	Position  Position // zero value means use the model default
}

// Resolve implements AlertSpec. AlertDefinition is already fully specified,
// so it returns itself regardless of font style.
func (d AlertDefinition) Resolve(_ FontStyle) AlertDefinition {
	return d
}

// AlertLevel is a font-style-indexed bundle of three AlertDefinitions — one
// for ASCII, Unicode, and NerdFont. The model's stored FontStyle determines
// which variant is used at alert time.
type AlertLevel struct {
	ascii    AlertDefinition
	unicode  AlertDefinition
	nerdFont AlertDefinition
}

// Resolve implements AlertSpec, picking the variant that matches f.
func (l AlertLevel) Resolve(f FontStyle) AlertDefinition {
	switch f {
	case FontUnicode:
		return l.unicode
	case FontNerdFont:
		return l.nerdFont
	default:
		return l.ascii
	}
}

// Built-in alert levels — one per severity. Pass to NewAlertCmd; the model
// resolves the correct variant based on the FontStyle given to New().
var (
	InfoAlert = AlertLevel{
		ascii:    AlertDefinition{Prefix: "(i)", ForeColor: "#00FF00"},
		unicode:  AlertDefinition{Prefix: "ⓘ", ForeColor: "#00FF00"},
		nerdFont: AlertDefinition{Prefix: "\U000f02fc", ForeColor: "#00FF00"},
	}

	WarnAlert = AlertLevel{
		ascii:    AlertDefinition{Prefix: "(!)", ForeColor: "#FFFF00"},
		unicode:  AlertDefinition{Prefix: "⚠", ForeColor: "#FFFF00"},
		nerdFont: AlertDefinition{Prefix: "󱈸", ForeColor: "#FFFF00"},
	}

	ErrorAlert = AlertLevel{
		ascii:    AlertDefinition{Prefix: "[!!]", ForeColor: "#FF0000"},
		unicode:  AlertDefinition{Prefix: "✘", ForeColor: "#FF0000"},
		nerdFont: AlertDefinition{Prefix: "󰬅", ForeColor: "#FF0000"},
	}

	DebugAlert = AlertLevel{
		ascii:    AlertDefinition{Prefix: "(?)", ForeColor: "#FF00FF"},
		unicode:  AlertDefinition{Prefix: "?", ForeColor: "#FF00FF"},
		nerdFont: AlertDefinition{Prefix: "󰃤", ForeColor: "#FF00FF"},
	}
)
