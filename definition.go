package toast

// FontStyle selects the glyph set used for alert prefixes.
type FontStyle string

const (
	FontASCII    FontStyle = "ascii"
	FontUnicode  FontStyle = "unicode"
	FontNerdFont FontStyle = "nerdfont"
)

// AlertDefinition holds the visual configuration for one alert level.
type AlertDefinition struct {
	Prefix    string
	ForeColor string
	Position  Position // zero value means use the model default
}

// Built-in alert definitions — one per level × font style.
var (
	InfoAlertASCII    = AlertDefinition{Prefix: "(i)", ForeColor: "#00FF00"}
	InfoAlertUnicode  = AlertDefinition{Prefix: "ⓘ", ForeColor: "#00FF00"}
	InfoAlertNerdFont = AlertDefinition{Prefix: "", ForeColor: "#00FF00"}

	WarnAlertASCII    = AlertDefinition{Prefix: "(!)", ForeColor: "#FFFF00"}
	WarnAlertUnicode  = AlertDefinition{Prefix: "⚠", ForeColor: "#FFFF00"}
	WarnAlertNerdFont = AlertDefinition{Prefix: "󱈸", ForeColor: "#FFFF00"}

	ErrorAlertASCII    = AlertDefinition{Prefix: "[!!]", ForeColor: "#FF0000"}
	ErrorAlertUnicode  = AlertDefinition{Prefix: "✘", ForeColor: "#FF0000"}
	ErrorAlertNerdFont = AlertDefinition{Prefix: "󰬅", ForeColor: "#FF0000"}

	DebugAlertASCII    = AlertDefinition{Prefix: "(?)", ForeColor: "#FF00FF"}
	DebugAlertUnicode  = AlertDefinition{Prefix: "?", ForeColor: "#FF00FF"}
	DebugAlertNerdFont = AlertDefinition{Prefix: "󰃤", ForeColor: "#FF00FF"}
)
