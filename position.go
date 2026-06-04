package toast

// Position specifies where on screen an Alert overlay appears.
type Position string

const (
	// UnspecifiedPosition is the zero value; callers must set a real position.
	UnspecifiedPosition Position = ""

	TopLeft      Position = "top-left"
	TopCenter    Position = "top-center"
	TopRight     Position = "top-right"
	BottomLeft   Position = "bottom-left"
	BottomCenter Position = "bottom-center"
	BottomRight  Position = "bottom-right"
)

// IsValid reports whether p is one of the six defined positions.
func (p Position) IsValid() bool {
	switch p {
	case TopLeft, TopCenter, TopRight, BottomLeft, BottomCenter, BottomRight:
		return true
	}
	return false
}
