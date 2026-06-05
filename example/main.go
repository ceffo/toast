package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ceffo/toast"
)

// SuccessAlert is a user-defined custom AlertDefinition (not a built-in).
var SuccessAlert = toast.AlertDefinition{
	Prefix:    "✓",
	ForeColor: "#00FF88",
	Position:  toast.TopCenter,
}

const (
	toastWidth   = 60
	minDuration  = 500 * time.Millisecond
	maxDuration  = 10 * time.Second
	durationStep = 500 * time.Millisecond
)

type model struct {
	toast    toast.Model
	width    int
	height   int
	duration time.Duration
	position toast.Position
}

func buildToast(duration time.Duration, position toast.Position) toast.Model {
	return toast.New(toastWidth, toast.FontNerdFont, duration).
		WithMinWidth(20).
		WithPosition(position).
		WithQueueDepth(5).
		WithAllowEscToClose()
}

func initialModel() model {
	dur := 2 * time.Second
	pos := toast.TopRight
	return model{
		toast:    buildToast(dur, pos),
		width:    80,
		height:   24,
		duration: dur,
		position: pos,
	}
}

func (m model) Init() tea.Cmd {
	return m.toast.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			// ESC guard: only quit when no alert is active; otherwise let
			// toast.Update consume it and close the current alert.
			if !m.toast.HasActiveAlert() {
				return m, tea.Quit
			}

		// 4 built-in alert types
		case "i":
			cmds = append(cmds, m.toast.NewAlertCmd(toast.InfoAlertUnicode, "Info: everything is running smoothly"))
		case "w":
			cmds = append(cmds, m.toast.NewAlertCmd(toast.WarnAlertUnicode, "Warn: disk usage is above 80%"))
		case "e":
			cmds = append(cmds, m.toast.NewAlertCmd(toast.ErrorAlertUnicode, "Error: connection refused on port 8080"))
		case "d":
			cmds = append(cmds, m.toast.NewAlertCmd(toast.DebugAlertUnicode, "Debug: goroutines=42 heap=128MB"))

		// Custom AlertDefinition
		case "s":
			cmds = append(cmds, m.toast.NewAlertCmd(SuccessAlert, "Success: deployment completed in 4.2s"))

		// Burst — fires 5 alerts at once to demonstrate WithQueueDepth
		case "b":
			for k := 1; k <= 5; k++ {
				k := k
				cmds = append(cmds, m.toast.NewAlertCmd(toast.InfoAlertUnicode, fmt.Sprintf("Burst alert #%d of 5", k)))
			}

		// Duration — + increases, - decreases (step: 500ms, range: 500ms–10s)
		case "+":
			if m.duration < maxDuration {
				m.duration += durationStep
				m.toast = buildToast(m.duration, m.position)
			}
		case "-":
			if m.duration > minDuration {
				m.duration -= durationStep
				m.toast = buildToast(m.duration, m.position)
			}

		// 6 position bindings
		case "1":
			m.position = toast.TopLeft
			m.toast = m.toast.WithPosition(m.position)
		case "2":
			m.position = toast.TopCenter
			m.toast = m.toast.WithPosition(m.position)
		case "3":
			m.position = toast.TopRight
			m.toast = m.toast.WithPosition(m.position)
		case "4":
			m.position = toast.BottomLeft
			m.toast = m.toast.WithPosition(m.position)
		case "5":
			m.position = toast.BottomCenter
			m.toast = m.toast.WithPosition(m.position)
		case "6":
			m.position = toast.BottomRight
			m.toast = m.toast.WithPosition(m.position)
		}
	}

	var toastCmd tea.Cmd
	m.toast, toastCmd = m.toast.Update(msg)
	if toastCmd != nil {
		cmds = append(cmds, toastCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) helpLine() string {
	return fmt.Sprintf(
		"i=info  w=warn  e=error  d=debug  s=custom  b=burst  "+
			"1=top-left  2=top-center  3=top-right  "+
			"4=bottom-left  5=bottom-center  6=bottom-right  "+
			"+/-=duration(%s)  esc/q=quit",
		m.duration,
	)
}

func (m model) View() tea.View {
	height := m.height
	if height < 5 {
		height = 24
	}

	rowCount := height - 3
	rows := make([]string, rowCount)
	for i := range rows {
		rows[i] = fmt.Sprintf("  Row %3d │ Lorem ipsum dolor sit amet, consectetur adipiscing elit.", i+1)
	}

	content := strings.Join(rows, "\n") + "\n\n" + m.helpLine()

	return tea.View{
		Content:   m.toast.Render(content),
		AltScreen: true,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
