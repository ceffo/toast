package toast

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type alertMsg struct {
	a alert
}

type tickMsg struct{}

// Model holds the toast overlay state.
type Model struct {
	queue           []alert
	maxDepth        int
	width           int
	minWidth        int
	duration        time.Duration
	position        Position
	font            FontStyle
	allowEscToClose bool
	style           lipgloss.Style
}

// New creates a Model. font is stored and used by NewAlertCmd to resolve the
// correct AlertDefinition variant when an AlertLevel is passed.
func New(width int, font FontStyle, duration time.Duration) Model {
	return Model{
		width:    width,
		font:     font,
		duration: duration,
		maxDepth: 5,
		style:    lipgloss.NewStyle(),
	}
}

func (m Model) WithMinWidth(min int) Model {
	m.minWidth = min
	return m
}

func (m Model) WithPosition(pos Position) Model {
	m.position = pos
	return m
}

func (m Model) WithQueueDepth(depth int) Model {
	m.maxDepth = depth
	return m
}

func (m Model) WithAllowEscToClose() Model {
	m.allowEscToClose = true
	return m
}

// Init returns nil — no startup command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

func tick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Update handles alertMsg, tickMsg, and esc key.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case alertMsg:
		a := msg.a
		a.deathTime = time.Now().Add(m.duration)
		wasEmpty := len(m.queue) == 0
		q := append([]alert{}, m.queue...)
		if m.maxDepth > 0 && len(q) >= m.maxDepth {
			q = q[1:]
		}
		m.queue = append(q, a)
		if wasEmpty {
			return m, tick()
		}
		return m, nil

	case tickMsg:
		if len(m.queue) == 0 {
			return m, nil
		}
		if time.Now().After(m.queue[0].deathTime) {
			m.queue = m.queue[1:]
			if len(m.queue) == 0 {
				return m, nil
			}
			return m, tick()
		}
		q := append([]alert{}, m.queue...)
		elapsed := m.duration - time.Until(q[0].deathTime)
		step := float64(elapsed) / float64(m.duration)
		if step < 0 {
			step = 0
		} else if step > 1 {
			step = 1
		}
		q[0].curLerpStep = step
		m.queue = q
		return m, tick()

	case tea.KeyPressMsg:
		if m.allowEscToClose && msg.String() == "esc" && len(m.queue) > 0 {
			m.queue = m.queue[1:]
			if len(m.queue) == 0 {
				return m, nil
			}
			return m, tick()
		}
	}
	return m, nil
}

// HasActiveAlert reports whether there are any queued alerts.
func (m Model) HasActiveAlert() bool {
	return len(m.queue) > 0
}

// NewAlertCmd returns a tea.Cmd that enqueues an alert.
// spec may be an AlertLevel (resolved using the model's FontStyle) or a fully
// specified AlertDefinition. Uses def.Position if valid, otherwise falls back
// to the model's position.
func (m Model) NewAlertCmd(spec AlertSpec, msg string) tea.Cmd {
	def := spec.Resolve(m.font)
	pos := def.Position
	if !pos.IsValid() {
		pos = m.position
	}
	foreColor, _ := colorful.Hex(def.ForeColor)
	a := alert{
		message:   msg,
		prefix:    def.Prefix,
		foreColor: foreColor,
		style:     m.style,
		width:     float64(m.width),
		minWidth:  float64(m.minWidth),
		position:  pos,
	}
	return func() tea.Msg {
		return alertMsg{a: a}
	}
}

// Render overlays the head alert on content and returns the composited string.
// Returns content unchanged if no alerts are queued.
func (m Model) Render(content string) string {
	if len(m.queue) == 0 {
		return content
	}

	head := m.queue[0]
	contentW := lipgloss.Width(content)
	alertMaxW := m.width
	if contentW < alertMaxW {
		alertMaxW = contentW
	}
	head.width = float64(alertMaxW)

	rendered := head.render()
	alertW := lipgloss.Width(rendered)
	alertH := lipgloss.Height(rendered)
	contentH := lipgloss.Height(content)

	x, y := alertCoords(head.position, contentW, contentH, alertW, alertH)

	bg := lipgloss.NewLayer(content)
	fg := lipgloss.NewLayer(rendered).X(x).Y(y)
	return lipgloss.NewCompositor(bg, fg).Render()
}

func alertCoords(pos Position, contentW, contentH, alertW, alertH int) (int, int) {
	var x, y int
	switch pos {
	case TopLeft:
		x, y = 0, 0
	case TopCenter:
		x, y = (contentW-alertW)/2, 0
	case TopRight:
		x, y = contentW-alertW, 0
	case BottomLeft:
		x, y = 0, contentH-alertH
	case BottomCenter:
		x, y = (contentW-alertW)/2, contentH-alertH
	case BottomRight:
		x, y = contentW-alertW, contentH-alertH
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}
